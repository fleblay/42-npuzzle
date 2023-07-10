package main

import (
	"container/heap"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

var evals = []eval{
	//{"dijkstra", dijkstra},
	//	{"greedy_hamming", greedy_hamming},
	//	{"greedy_inv", greedy_inv},
	//{"greedy_manhattan", greedy_manhattan},
	{"astar_manhattan", astar_manhattan_generator(1)},
	//{"astar_manhattan", astar_manhattan_generator(2)},
	//	{"astar_hamming", astar_hamming},
	//{"astar_inversion", astar_inv},
}

type safeData struct {
	mu           sync.Mutex
	muQueue      []sync.Mutex
	posQueue     []*PriorityQueue
	seenNodes    map[string]int
	tries        int
	maxSizeQueue int
	path         []byte // solution
	over         bool
	end          chan bool
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

func printTimeInfo(elapsed []time.Duration) {

	fmt.Println("time to get next node to explore from Queue :", elapsed[0].String())
	fmt.Println("time applying moves:", elapsed[1].String())
	fmt.Println("time calculating costs and creating node:", elapsed[2].String())
	fmt.Println("time finding if node already exists in nodes list:", elapsed[3].String())
	fmt.Println("time adding node to queues:", elapsed[4].String())
}

func getNextNodeIndex(queue []Node) int {
	retScore := queue[0].score
	ret := 0
	for index, value := range queue {
		if value.score < retScore {
			retScore = value.score
			ret = index
		}
	}
	return ret
}

func getNextMoves(startPos, goalPos [][]int, scoreFx evalFx, path []byte, currentNode *Item, elapsed []time.Duration, data *safeData, index int, workers int) {
	for indexDir, dir := range directions {
		start := time.Now()
		ok, nextPos := dir.fx(currentNode.node.world)
		end := time.Now()
		elapsed[1] += end.Sub(start)
		if !ok {
			continue
		}
		start = time.Now()
		score := scoreFx(nextPos, startPos, goalPos, path)
		nextPath := DeepSliceCopyAndAdd(path, dir.name)
		nextNode := Node{nextPos, nextPath, score}
		end = time.Now()
		elapsed[2] += end.Sub(start)
		start = time.Now()
		keyNode := matrixToString(nextPos)
		data.mu.Lock()
		seenNodesScore, alreadyExplored := (data.seenNodes)[keyNode]
		data.mu.Unlock()
		end = time.Now()
		elapsed[3] += end.Sub(start)
		if !alreadyExplored ||
			score < seenNodesScore {
			start = time.Now()
			item := &Item{node: nextNode}
			queueIndex := (index + indexDir) % workers
			data.muQueue[queueIndex].Lock()
			//fmt.Println("Push [1] :index is : ", index, "len of queue :", len(*data.posQueue[index]))
			heap.Push(data.posQueue[queueIndex], item)
			//fmt.Println("Push [1] :index is : ", index, "len of queue :", len(*data.posQueue[index]))
			data.muQueue[queueIndex].Unlock()
			data.mu.Lock()
			(data.seenNodes)[keyNode] = score
			data.mu.Unlock()
			end = time.Now()
			elapsed[4] += end.Sub(start)
		}
	}
}

func algo(world [][]int, scoreFx evalFx, data *safeData, index int, workers int) {
	goalPos := goal(len(world))
	startPos := Deep2DSliceCopy(world)
	var elapsed [8]time.Duration
	startAlgo := time.Now()
	for {
		data.mu.Lock()
		data.tries++
		tries := data.tries
		over := data.over
		data.muQueue[index].Lock()
		lenqueue := len(*data.posQueue[index])
		data.maxSizeQueue = Max(data.maxSizeQueue, lenqueue)
		if lenqueue == 0 {
			data.muQueue[index].Unlock()
			data.mu.Unlock()
			fmt.Println(index, "Empty queue. Waiting")
			time.Sleep(10 * time.Millisecond)
			continue
		} else if over {
			data.muQueue[index].Unlock()
			data.mu.Unlock()
			fmt.Println(index, "End of sim")
			return
		}
		start := time.Now()
		//fmt.Println("Pop : [1] :index is : ", index, "len of queue :", len(*data.posQueue[index]))
		currentNode := (heap.Pop(data.posQueue[index])).(*Item) // Parfois erreur ????
		//fmt.Println("Pop [2] : index is : ", index, "len of queue :", len(*data.posQueue[index]))
		data.muQueue[index].Unlock()
		data.mu.Unlock()
		end := time.Now()
		elapsed[0] += end.Sub(start)

		currentPath := currentNode.node.path
		if tries > 0 && tries%100000 == 0 {
			fmt.Printf("[%d] Time so far : %s | %d * 100k tries. Len of try : %d. Score : %d Len of Queue : %d\n", index, time.Since(startAlgo), tries/100000, len(currentNode.node.path), currentNode.node.score, lenqueue)
		}

		if isEqual(goalPos, currentNode.node.world) {
			printTimeInfo(elapsed[:])
			data.mu.Lock()
			data.path = currentPath
			data.over = true
			data.end <- true
			data.mu.Unlock()
			return
		}
		getNextMoves(startPos, goalPos, scoreFx, currentPath, currentNode, elapsed[:], data, index, workers)
	}
}

func initData(board [][]int, worker int) (data *safeData) {
	data = &safeData{}
	data.seenNodes = make(map[string]int, 1000000)
	startPos := Deep2DSliceCopy(board)
	data.seenNodes[matrixToString(startPos)] = 0
	data.posQueue = make([]*PriorityQueue, worker)
	for i := 0; i < worker; i++ {

		queue := make(PriorityQueue, 1, 1000000)
		queue[0] = &Item{node: Node{world: startPos, score: 0, path: []byte{}}}
		data.posQueue[i] = &queue
		heap.Init(data.posQueue[i])
	}
	data.end = make(chan bool)
	data.over = false
	data.muQueue = make([]sync.Mutex, worker)
	return
}

func main() {
	var (
		file      string
		mapSize   int
		heuristic string
	)
	flag.StringVar(&file, "f", "", "usage : -f [filename]")
	flag.IntVar(&mapSize, "s", 3, "usage : -s [size]")
	flag.StringVar(&heuristic, "h", "m", "usage : -h m for manhattan or e for euclidean")
	flag.Parse()

	var board [][]int

	if file != "" {
		file := OpenFile(file)
		_, board = ParseInput(file)
	} else if mapSize > 0 {
		board = gridGenerator(mapSize)
	} else {
		fmt.Println("Invalid Map size")
		os.Exit(1)
	}
	if !isSolvable(board) {
		fmt.Println("Board is not solvable")
		os.Exit(1)
	}
	fmt.Println("Board is :", board)
	for _, eval := range evals {
		fmt.Println("Now starting with :", eval.name)
		start := time.Now()
		workers := 4
		data := initData(board, workers)
		for i := 0; i < workers; i++ {
			go algo(board, eval.fx, data, i, workers)
		}
		<-data.end
		end := time.Now()
		elapsed := end.Sub(start)
		if data.path != nil {
			//displayBoard(board, path, seenPos, eval.name+" in "+elapsed.String(), tries, sizeMax)
			fmt.Println("Succes with :", eval.name, "in ", elapsed.String(), "!")
			fmt.Printf("len of solution %v, %d pos seen, %d tries, %d space complexity\n", len(data.path), len(data.seenNodes), data.tries, data.maxSizeQueue)
			fmt.Println(string(data.path))
		} else {
			fmt.Println("No solution !")
		}
	}
	/*
		for playBoard(board) {
			mapSize = 3
			board = gridGenerator(mapSize)
		}
	*/
}
