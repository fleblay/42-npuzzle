package algo

import (
	"container/heap"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"sync"
	"time"
)

func initData(param AlgoParameters) (data safeData) {
	startPos := param.Board
	data.SeenNodes = make([]map[string]int, param.SeenNodesSplit)
	keyNode, _, _ := MatrixToStringSelector(startPos, param.Workers, param.SeenNodesSplit)
	for i := 0; i < param.SeenNodesSplit; i++ {
		data.SeenNodes[i] = make(map[string]int, 1000)
		data.SeenNodes[i][keyNode] = 0
	}
	data.PosQueue = make([]*PriorityQueue, param.Workers)
	for i := 0; i < param.Workers; i++ {

		queue := make(PriorityQueue, 1, 1000)
		queue[0] = &Item{node: Node{world: startPos, score: 0, path: []byte{}}}
		data.PosQueue[i] = &queue
		heap.Init(data.PosQueue[i])
	}
	data.Over = false
	data.Win = false
	data.MuQueue = make([]sync.Mutex, param.Workers)
	data.MuSeen = make([]sync.Mutex, param.SeenNodesSplit)
	data.MaxSizeQueue = make([]int, param.Workers)
	data.Idle = 0
	currentAvailableRAM, _ := GetAvailableRAM()
	if (param.RAMMaxGB << 30) < currentAvailableRAM {
		data.RAMMin = currentAvailableRAM - (param.RAMMaxGB << 30)
		fmt.Fprintln(os.Stderr, "RAM Min left for system is now :", data.RAMMin >> 20, "MB")
	} else {
		fmt.Fprintln(os.Stderr, "Max Ram Usage specified is superior to current available RAM. Setting RAM Min to 0. Default value will prevail")
		data.RAMMin = 0
	}
	return
}

func launchAstarWorkers(param AlgoParameters, data *safeData) (result Result) {
	fmt.Fprintln(os.Stderr, "Selected ALGO : A*")
	var wg sync.WaitGroup
	for i := 0; i < param.Workers; i++ {
		wg.Add(1)
		go func(param AlgoParameters, data *safeData, i int) {

			algo(param, data, i)
			wg.Done()
		}(param, data, i)
	}
	wg.Wait()
	min, max, indexmin, indexmax := 1<<31, 0, -1, -1
	for index, value := range data.SeenNodes {
		currLen := len(value)
		data.ClosedSetComplexity += currLen
		if currLen > max {
			max = currLen
			indexmax = index
		}
		if currLen < min {
			min = currLen
			indexmin = index
		}
	}
	fmt.Fprintf(os.Stderr, "NodePool max count difference : %d k for [%d] - [%d]. Mean : %d k\n", (max-min)/1000, indexmax, indexmin, data.ClosedSetComplexity/(1000*len(data.SeenNodes)))
	switch {
	case data.Win == true:
		fmt.Fprintln(os.Stderr, "Found a solution")
	case data.RamFailure == true:
		fmt.Fprintln(os.Stderr, "RAM Failure")
	}
	return Result{data.Path, data.ClosedSetComplexity, data.Tries, data.RamFailure, "A*"}
}

func checkOptimalSolution(currentNode *Item, data *safeData) bool {
	bestNodes := make([]*Item, len(data.PosQueue))
	for i := range data.PosQueue {
		data.MuQueue[i].Lock()
		if data.PosQueue[i].Len() > 0 {
			bestNodes[i] = heap.Pop(data.PosQueue[i]).(*Item)
		} else {
			bestNodes[i] = nil
		}
		data.MuQueue[i].Unlock()
	}
	//Check if at least one of the nodes has a score strictly inferior
	for i := range bestNodes {
		if bestNodes[i] != nil && bestNodes[i].node.score < currentNode.node.score {
			for j := range bestNodes {
				data.MuQueue[j].Lock()
				if bestNodes[j] != nil {
					heap.Push(data.PosQueue[j], bestNodes[j])
				}
				data.MuQueue[j].Unlock()
			}
			return false
		}
	}

	return true
}

func algo(param AlgoParameters, data *safeData, workerIndex int) {
	goalPos := Goal(len(param.Board))
	startPos := param.Board
	var foundSol *Item
	startAlgo := time.Now()
	isIdle := false
	for {
		over, tries, lenqueue, idle, ramFailure := refreshData(data, workerIndex)
		if ramFailure {
			fmt.Fprintf(os.Stderr, "[%2d] - Someone had a ram failure. Leaving now\n", workerIndex)
			return
		}
		if over {
			fmt.Fprintf(os.Stderr, "[%2d] - Someone ended sim. Leaving now\n", workerIndex)
			return
		}
		if idle >= param.Workers {
			fmt.Fprintf(os.Stderr, "[%2d] - Everyone is idle\n", workerIndex)
			return
		}
		if isIdle && lenqueue > 0 {
			isIdle = false
			data.Mu.Lock()
			data.Idle--
			data.Mu.Unlock()
		}
		if lenqueue == 0 {
			if !isIdle {
				data.Mu.Lock()
				data.Idle++
				data.Mu.Unlock()
				isIdle = true
			}
			continue
		}
		currentNode := getNextNode(data, workerIndex)
		if currentNode == nil {
			continue
		}
		if foundSol != nil && currentNode.node.score >= foundSol.node.score {
			data.Mu.Lock()
			fmt.Fprintf(os.Stderr, "\x1b[32m[%2d] - Found an OPTIMAL solution\n\x1b[0m", workerIndex)
			terminateSearch(data, foundSol.node.path, foundSol.node.score)
			data.Mu.Unlock()
			return
		}
		printInfo(workerIndex, tries, currentNode, startAlgo, lenqueue)
		if isEqual(goalPos, currentNode.node.world) {
			data.Mu.Lock()
			if checkOptimalSolution(currentNode, data) {
				fmt.Fprintf(os.Stderr, "\x1b[32m[%2d] - Found an OPTIMAL solution\n\x1b[0m", workerIndex)
				terminateSearch(data, currentNode.node.path, currentNode.node.score)
				data.Mu.Unlock()
				return
			} else {
				fmt.Fprintf(os.Stderr, "\x1b[33m[%2d] - Found a solution : Caching result\n\x1b[0m", workerIndex)
				foundSol = currentNode
				data.Mu.Unlock()
			}
		}
		if tries%1000 == 0 {
			availableRAM, err := GetAvailableRAM()
			if err != nil ||
				availableRAM>>20 < MinRAMAvailableMB ||
				availableRAM < data.RAMMin {
				fmt.Fprintf(os.Stderr, "[%d] - Not enough RAM[%v MB] to continue or Fatal (error reading RAM status)\n", workerIndex, availableRAM>>20)
				data.Mu.Lock()
				data.RamFailure = true
				data.Mu.Unlock()
				continue
			}
		}
		getNextMoves(startPos, goalPos, param.Eval.Fx, currentNode.node.path, currentNode, data, workerIndex, param.Workers, param.SeenNodesSplit)
	}
}

func terminateSearch(data *safeData, solutionPath []byte, score int) {
	data.Path = solutionPath
	data.Over = true
	data.Win = true
	data.WinScore = score
}

func GetAvailableRAM() (uint64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("Error while getting info about memory: %v", err)
	}
	availableRAM := v.Available
	return availableRAM, nil
}

func getNextMoves(startPos, goalPos [][]int, scoreFx EvalFx, path []byte, currentNode *Item, data *safeData, index int, workers int, seenNodesSplit int) {
	for _, dir := range Directions {
		if len(path) > 0 {
			conflictStr := string(path[len(path)-1]) + string(dir.name)
			if conflictStr == "LR" || conflictStr == "RL" || conflictStr == "UD" || conflictStr == "DU" {
				continue
			}
		}
		ok, nextPos := dir.fx(currentNode.node.world)
		if !ok {
			continue
		}
		score := scoreFx(nextPos, startPos, goalPos, path)
		nextPath := DeepSliceCopyAndAdd(path, dir.name)
		nextNode := Node{world: nextPos, path: nextPath, score: score}
		keyNode, queueIndex, seenNodeIndex := MatrixToStringSelector(nextPos, workers, seenNodesSplit)
		data.MuSeen[seenNodeIndex].Lock()
		seenNodesScore, alreadyExplored := data.SeenNodes[seenNodeIndex][keyNode]
		data.MuSeen[seenNodeIndex].Unlock()
		if !alreadyExplored ||
			score < seenNodesScore {
			item := &Item{node: nextNode}
			data.MuQueue[queueIndex].Lock()
			heap.Push(data.PosQueue[queueIndex], item)
			data.MuQueue[queueIndex].Unlock()
			data.MuSeen[seenNodeIndex].Lock()
			data.SeenNodes[seenNodeIndex][keyNode] = score
			data.MuSeen[seenNodeIndex].Unlock()
		}
	}
}

func refreshData(data *safeData, workerIndex int) (over bool, tries, lenqueue int, idle int, ramFailure bool) {
	data.Mu.Lock()

	data.Tries++
	tries = data.Tries
	over = data.Over
	idle = data.Idle
	ramFailure = data.RamFailure
	data.Mu.Unlock()
	data.MuQueue[workerIndex].Lock()
	lenqueue = len(*data.PosQueue[workerIndex])
	data.MaxSizeQueue[workerIndex] = Max(data.MaxSizeQueue[workerIndex], lenqueue)
	data.MuQueue[workerIndex].Unlock()
	return
}

func printInfo(workerIndex int, tries int, currentNode *Item, startAlgo time.Time, lenqueue int) {
	if tries > 0 && tries%100000 == 0 {
		fmt.Fprintf(os.Stderr, "[%2d] Time so far : %s | %d * 100k tries. Len of try : %d. Score : %d Len of Queue : %d\n", workerIndex, time.Since(startAlgo), tries/100000, len(currentNode.node.path), currentNode.node.score, lenqueue)
	}
}

func getNextNode(data *safeData, workerIndex int) (currentNode *Item) {
	data.MuQueue[workerIndex].Lock()
	if data.PosQueue[workerIndex].Len() != 0 {
		currentNode = (heap.Pop(data.PosQueue[workerIndex])).(*Item)
	}
	data.MuQueue[workerIndex].Unlock()
	return
}
