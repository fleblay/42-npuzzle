package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"github.com/shirou/gopsutil/v3/mem"
)

func areFlagsOk(workers int, seenNodesSplit int, heuristic string, mapSize int) (fx eval, ok bool) {
	if workers < 1 || workers > 16 {
		fmt.Println("Invalid number of workers")
		os.Exit(1)
	}
	if seenNodesSplit < 1 || seenNodesSplit > 64 {
		fmt.Println("Invalid number of splits")
		os.Exit(1)
	}
	if mapSize < 3 || mapSize > 10 {
		fmt.Println("Invalid map size")
		os.Exit(1)
	}
	for _, current := range evals {
		if current.name == heuristic {
			return current, true
		}
	}
	fmt.Println("Invalid heuristic")
	return eval{}, false
}

func handleSignals() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-sigc
		fmt.Println("\b\bExiting after receiving a signal")
		os.Exit(1)
	}()
}

func getAvailableRAM() (uint64, error) {
	v, err :=  mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("Error while getting info about memory: %v", err)
	}
	availableRAM := v.Free
	return availableRAM, nil
}

func iterateAlgo(board [][]int, maxScore int, workers int, seenNodesSplit int, evalfx evalFx, data *safeData) {
	var wg sync.WaitGroup
Iteration:
	for maxScore < 1<<31 {
		fmt.Fprintln(os.Stderr, "cut off is now :", maxScore)
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(board [][]int, evalfx evalFx, data *safeData, i int, workers int, seenNodesSplit int, maxScore int) {

				algo(board, evalfx, data, i, workers, seenNodesSplit, maxScore)
				wg.Done()
			}(board, evalfx, data, i, workers, seenNodesSplit, maxScore)
		}
		wg.Wait()
		switch data.win {
		case true:
			fmt.Fprintln(os.Stderr, "Found a solution")
			break Iteration
		default:
			*data = initData(board, workers, seenNodesSplit)
			maxScore +=2
		}
	}
}

func main() {
	var (
		file                string
		mapSize             int
		heuristic           string
		workers             int
		seenNodesSplit      int
		speedDisplay        int
		noIterativeDepth    bool
		debug               bool
		disableUI           bool
		maxScore            int
		closedSetComplexity int
		board               [][]int
	)

	flag.StringVar(&file, "f", "", "usage : -f [filename]")
	flag.IntVar(&mapSize, "s", 3, "usage : -s [size]")
	flag.StringVar(&heuristic, "h", "astar_manhattan", "usage : -h [heuristic]")
	flag.IntVar(&workers, "w", 1, "usage : -w [workers] between 1 and 16")
	flag.IntVar(&seenNodesSplit, "split", 1, "usage : -split [setNodesSplit] between 1 and 32")
	flag.IntVar(&speedDisplay, "speed", 100, "usage : -speed [speedDisplay] between 1 and 1000")
	flag.BoolVar(&noIterativeDepth, "no-i", false, "usage : -no-i")
	flag.BoolVar(&debug, "d", false, "usage : -d")
	flag.BoolVar(&disableUI, "no-ui", false, "usage : -no-ui")
	flag.Parse()

	eval, ok := areFlagsOk(workers, seenNodesSplit, heuristic, mapSize)
	if !ok {
		os.Exit(1)
	}
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
		fmt.Println("No valid map size or filename option missing")
		os.Exit(1)
	}
	if !isSolvable(board) {
		fmt.Println("Board is not solvable")
		if !disableUI {
			displayBoard(board, []byte{}, eval.name, "", 0, 0, workers, seenNodesSplit, speedDisplay)
		}
		os.Exit(0)
	}
	fmt.Println("Board is :", board)
	fmt.Println("Now starting with :", eval.name)
	data := initData(board, workers, seenNodesSplit)
	if !noIterativeDepth {
		fmt.Println("Search Method : IDA*")
		maxScore = eval.fx(board, board, goal(len(board)), []byte{}) + 1
	} else {
		fmt.Println("Search Method : A*")
		maxScore |= (1<<31 - 1)
	}
	handleSignals()
	start := time.Now()
	iterateAlgo(board, maxScore, workers, seenNodesSplit, eval.fx, &data)
	end := time.Now()
	elapsed := end.Sub(start)
	if data.path != nil {
		for _, value := range data.seenNodes {
			closedSetComplexity += len(value)
		}
		fmt.Println("Succes with :", eval.name, "in ", elapsed.String(), "!")
		fmt.Printf("len of solution : %v, time complexity / tries : %d, space complexity : %d, score : %d\n", len(data.path), data.tries, closedSetComplexity, data.winScore)
		if !disableUI {
			displayBoard(board, data.path, eval.name, elapsed.String(), data.tries, closedSetComplexity, workers, seenNodesSplit, speedDisplay)
		}
		fmt.Println(string(data.path))
	} else {
		fmt.Println("No solution !")
	}
}
