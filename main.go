package main

import (
	"container/heap"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var evals = []eval{
	{"dijkstra", dijkstra},
	{"greedy_hamming", greedy_hamming},
	{"greedy_inv", greedy_inv},
	{"greedy_manhattan", greedy_manhattan},
	{"astar_manhattan", astar_manhattan_generator(1)},
	{"astar_manhattan2", astar_manhattan_generator(2)},
	{"astar_inversion", astar_inv},
}

type safeData struct {
	mu sync.Mutex

	muQueue  []sync.Mutex
	posQueue []*PriorityQueue

	muSeen    []sync.Mutex
	seenNodes []map[string]int

	tries        int
	maxSizeQueue []int

	path     []byte
	over     bool
	win      bool
	winScore int
	idle     int
	//end  chan bool
}

var directions = []struct {
	name byte
	fx   moveFx
}{
	{'U', moveUp},
	{'D', moveDown},
	{'L', moveLeft},
	{'R', moveRight},
}

func terminateSearch(data *safeData, solutionPath []byte, score int) {
	data.path = solutionPath
	data.over = true
	data.win = true
	data.winScore = score
}

func getNextMoves(startPos, goalPos [][]int, scoreFx evalFx, path []byte, currentNode *Item, data *safeData, index int, workers int, seenNodesSplit int, maxScore int) {
	for _, dir := range directions {
		ok, nextPos := dir.fx(currentNode.node.world)
		if !ok {
			continue
		}
		score := scoreFx(nextPos, startPos, goalPos, path)
		nextPath := DeepSliceCopyAndAdd(path, dir.name)
		nextNode := Node{world: nextPos, path: nextPath, score: score}
		keyNode, queueIndex, seenNodeIndex := matrixToStringSelector(nextPos, workers, seenNodesSplit)
		data.muSeen[seenNodeIndex].Lock()
		seenNodesScore, alreadyExplored := data.seenNodes[seenNodeIndex][keyNode]
		data.muSeen[seenNodeIndex].Unlock()
		if !alreadyExplored ||
			score < seenNodesScore {
			item := &Item{node: nextNode}
			if score < maxScore {
				data.muQueue[queueIndex].Lock()
				heap.Push(data.posQueue[queueIndex], item)
				data.muQueue[queueIndex].Unlock()
			}
			data.muSeen[seenNodeIndex].Lock()
			data.seenNodes[seenNodeIndex][keyNode] = score
			data.muSeen[seenNodeIndex].Unlock()
		}
	}
}

func noMoreNodesToExplore(data *safeData) bool {
	data.mu.Lock()
	totalLen := 0
	for i := range data.posQueue {
		data.muQueue[i].Lock()
		length := data.posQueue[i].Len()
		totalLen += length
		data.muQueue[i].Unlock()
	}
	data.mu.Unlock()
	if totalLen == 0 {
		fmt.Fprintln(os.Stderr, "all queues are empty. Leaving")
		return true
	} else {
		return false
	}
}

func refreshData(data *safeData, workerIndex int) (over bool, tries, lenqueue int, idle int) {
	data.mu.Lock()

	data.tries++
	tries = data.tries
	over = data.over
	idle = data.idle
	data.mu.Unlock()
	data.muQueue[workerIndex].Lock()
	lenqueue = len(*data.posQueue[workerIndex])
	data.maxSizeQueue[workerIndex] = Max(data.maxSizeQueue[workerIndex], lenqueue)
	data.muQueue[workerIndex].Unlock()

	return
}

func printInfo(workerIndex int, tries int, currentNode *Item, startAlgo time.Time, lenqueue int) {

	if tries > 0 && tries%100000 == 0 {
		fmt.Fprintf(os.Stderr, "[%2d] Time so far : %s | %d * 100k tries. Len of try : %d. Score : %d Len of Queue : %d\n", workerIndex, time.Since(startAlgo), tries/100000, len(currentNode.node.path), currentNode.node.score, lenqueue)
	}

}

func getNextNode(data *safeData, workerIndex int) (currentNode *Item) {
	data.muQueue[workerIndex].Lock()
	if data.posQueue[workerIndex].Len() != 0 {
		currentNode = (heap.Pop(data.posQueue[workerIndex])).(*Item)
	}
	data.muQueue[workerIndex].Unlock()
	return
}

func algo(world [][]int, scoreFx evalFx, data *safeData, workerIndex int, workers int, seenNodesSplit int, maxScore int) {
	goalPos := goal(len(world))
	startPos := Deep2DSliceCopy(world)
	var foundSol *Item
	startAlgo := time.Now()
	isIdle := false
	for {
		over, tries, lenqueue, idle := refreshData(data, workerIndex)
		if over {
			fmt.Fprintf(os.Stderr, "[%2d] - Someone ended sim. Leaving now\n", workerIndex)
			return
		}
		if idle >= workers {
			fmt.Fprintf(os.Stderr, "[%2d] - Everyone is idle\n", workerIndex)
			return
		}
		if isIdle && lenqueue > 0 {
			isIdle = false
			data.mu.Lock()
			data.idle--
			data.mu.Unlock()
		}
		if lenqueue == 0 {
			if !isIdle {
				data.mu.Lock()
				data.idle++
				data.mu.Unlock()
				isIdle = true
			}
			continue
		}
		currentNode := getNextNode(data, workerIndex)
		if currentNode == nil {
			continue
		}
		if foundSol != nil && currentNode.node.score > foundSol.node.score {
			data.mu.Lock()
			terminateSearch(data, foundSol.node.path, foundSol.node.score)
			data.mu.Unlock()
			return
		}
		printInfo(workerIndex, tries, currentNode, startAlgo, lenqueue)
		if isEqual(goalPos, currentNode.node.world) {
			data.mu.Lock()
			if checkOptimalSolution(currentNode, data) {
				fmt.Fprintf(os.Stderr, "\x1b[32m[%2d] - Found an OPTIMAL solution\n\x1b[0m", workerIndex)
				terminateSearch(data, currentNode.node.path, currentNode.node.score)
				data.mu.Unlock()
				return
			} else {
				fmt.Fprintf(os.Stderr, "\x1b[33m[%2d] - Found a NON optimal solution\n\x1b[0m", workerIndex)
				foundSol = currentNode
				data.mu.Unlock()
			}
		}
		getNextMoves(startPos, goalPos, scoreFx, currentNode.node.path, currentNode, data, workerIndex, workers, seenNodesSplit, maxScore)
	}
}

func checkOptimalSolution(currentNode *Item, data *safeData) bool {
	bestNodes := make([]*Item, len(data.posQueue))
	for i := range data.posQueue {
		data.muQueue[i].Lock()
		if data.posQueue[i].Len() > 0 {
			bestNodes[i] = heap.Pop(data.posQueue[i]).(*Item)
		} else {
			bestNodes[i] = nil
		}
		data.muQueue[i].Unlock()
	}
	for i := range bestNodes {
		if bestNodes[i] != nil && bestNodes[i].node.score <= currentNode.node.score {
			for j := range bestNodes {
				data.muQueue[j].Lock()
				if bestNodes[j] != nil {
					heap.Push(data.posQueue[j], bestNodes[j])
				}
				data.muQueue[j].Unlock()
			}
			return false
		}
	}

	return true
}

func initData(board [][]int, workers int, seenNodesSplit int) (data *safeData) {
	data = &safeData{}
	startPos := Deep2DSliceCopy(board)
	data.seenNodes = make([]map[string]int, seenNodesSplit)
	keyNode, _, _ := matrixToStringSelector(startPos, workers, seenNodesSplit)
	for i := 0; i < seenNodesSplit; i++ {
		data.seenNodes[i] = make(map[string]int, 1000)
		data.seenNodes[i][keyNode] = 0
	}
	data.posQueue = make([]*PriorityQueue, workers)
	for i := 0; i < workers; i++ {

		queue := make(PriorityQueue, 1, 1000)
		queue[0] = &Item{node: Node{world: startPos, score: 0, path: []byte{}}}
		data.posQueue[i] = &queue
		heap.Init(data.posQueue[i])
	}
	//data.end = make(chan bool)
	data.over = false
	data.win = false
	data.muQueue = make([]sync.Mutex, workers)
	data.muSeen = make([]sync.Mutex, seenNodesSplit)
	data.maxSizeQueue = make([]int, workers)
	data.idle = 0
	return
}

func checkFlags(workers int, seenNodesSplit int, heuristic string, mapSize int) eval {
	if workers < 1 || workers > 16 {
		fmt.Println("Invalid number of workers")
		os.Exit(1)
	}
	if seenNodesSplit < 1 || seenNodesSplit > 32 {
		fmt.Println("Invalid number of seenNodesSplit")
		os.Exit(1)
	}
	if mapSize < 3 || mapSize > 10 {
		fmt.Println("Invalid map size")
		os.Exit(1)
	}
	for _, current := range evals {
		if current.name == heuristic {
			return current
		}
	}

	fmt.Println("Invalid heuristic")
	os.Exit(1)
	return eval{}
}

func main() {
	var (
		file           string
		mapSize        int
		heuristic      string
		workers        int
		seenNodesSplit int
		speedDisplay   int
		iterativeDepth bool
		debug          bool
		ui             bool
	)
	flag.StringVar(&file, "f", "", "usage : -f [filename]")
	flag.IntVar(&mapSize, "s", 3, "usage : -s [size]")
	flag.StringVar(&heuristic, "h", "astar_manhattan", `usage : -h [heuristic] 
	- dijkstra
	- greedy_hamming
	- greedy_inv
	- greedy_manhattan
	- astar_manhattan
	- astar_manhattan2
	- astar_inversion
	`)
	flag.IntVar(&workers, "w", 1, "usage : -w [workers] between 1 and 16")
	flag.IntVar(&seenNodesSplit, "ss", 1, "usage : -ss [setNodesSplit] between 1 and 32")
	flag.IntVar(&speedDisplay, "sd", 100, "usage : -sd [speedDisplay] between 1 and 1000")
	flag.BoolVar(&iterativeDepth, "i", true, "usage : -i [true|false]")
	flag.BoolVar(&debug, "d", false, "usage : -d [true|false]")
	flag.BoolVar(&ui, "ui", false, "usage : -ui [true|false]")
	flag.Parse()

	var board [][]int
	eval := checkFlags(workers, seenNodesSplit, heuristic, mapSize)

	if !debug {
		newstderr, _ := os.Open("/dev/null")
		defer newstderr.Close()
		os.Stderr = newstderr
	}

	if file != "" {
		file := OpenFile(file)
		fmt.Println("Opening user provided map in file", file.Name())
		_, board = ParseInput(file)
	} else if mapSize > 0 {
		fmt.Println("Generating a map with size", mapSize)
		board = gridGenerator(mapSize)
	} else {
		fmt.Println("Invalid Map size")
		os.Exit(1)
	}
	if !isSolvable(board) && ui {
		fmt.Println("Board is not solvable")
		displayBoard(board, []byte{}, eval.name, "", 0, 0, workers, seenNodesSplit, speedDisplay)
		os.Exit(0)
	}
	fmt.Println("Board is :", board)

	fmt.Println("Now starting with :", eval.name)
	start := time.Now()
	data := initData(board, workers, seenNodesSplit)
	var maxScore int
	if iterativeDepth {
		fmt.Println("Search Method : A*")
		maxScore = eval.fx(board, board, goal(len(board)), []byte{}) + 1
	} else {
		fmt.Println("Search Method : IDA*")
		maxScore |= (1<<31 - 1)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGKILL)

	var wg sync.WaitGroup
Iteration:
	for maxScore < 1<<31 {
		fmt.Fprintln(os.Stderr, "cut off is now :", maxScore)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(board [][]int, evalfx evalFx, data *safeData, i int, workers int, seenNodeSplit int, maxScore int) {

				algo(board, evalfx, data, i, workers, seenNodeSplit, maxScore)
				wg.Done()
			}(board, eval.fx, data, i, workers, seenNodesSplit, maxScore)
		}
		wg.Wait()
		switch data.win {
		case true:
			fmt.Fprintln(os.Stderr, "Found a solution")
			break Iteration
		default:
			data = initData(board, workers, seenNodesSplit)
			maxScore++
		}
	}
	end := time.Now()
	elapsed := end.Sub(start)
	if data.path != nil {
		closedSetComplexity := 0
		for _, value := range data.seenNodes {
			closedSetComplexity += len(value)
		}

		fmt.Println("Succes with :", eval.name, "in ", elapsed.String(), "!")
		fmt.Printf("len of solution %v, %d time complexity / tries, %d space complexity, score : %d\n", len(data.path), data.tries, closedSetComplexity, data.winScore)
		if ui {
			displayBoard(board, data.path, eval.name, elapsed.String(), data.tries, closedSetComplexity, workers, seenNodesSplit, speedDisplay)
		}
		fmt.Println(string(data.path))
	} else {
		fmt.Println("No solution !")
	}

	/*
		for playBoard(board) {
			mapSize = 3
			board = gridGenerator(mapSize)
		}
	*/
}
