package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

func handleFatalError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error :", err.Error())
		os.Exit(1)
	}
}

func areFlagsOk(opt *option) (err error) {
	if opt.workers < 1 || opt.workers > 16 {
		return errors.New("Invalid number of workers")
	}
	if opt.seenNodesSplit < 1 || opt.seenNodesSplit > 256 {
		return errors.New("Invalid number of splits")
	}
	if opt.filename == "" && (opt.mapSize < 3 || opt.mapSize > 10) {
		return errors.New("Invalid map size")
	}
	for _, current := range evals {
		if current.name == opt.heuristic {
			return nil
		}
	}
	return errors.New("Invalid heuristic")
}

func setParam(opt *option, param *algoParameters) (err error) {
	param.workers = opt.workers
	param.seenNodesSplit = opt.seenNodesSplit
	for _, current := range evals {
		if current.name == opt.heuristic {
			param.eval = current
			break
		}
	}
	if !opt.debug {
		newstderr, err := os.Open("/dev/null")
		if err != nil {
			return err
		}
		defer newstderr.Close()
		os.Stderr = newstderr
	}
	if opt.filename != "" {
		fmt.Fprintln(os.Stderr, "Opening user provided map in file", opt.filename)
		opt.fd, err = OpenFile(opt.filename)
		if err != nil {
			return err
		}
		param.board = ParseInput(opt.fd)
	} else if opt.mapSize > 0 {
		fmt.Fprintln(os.Stderr, "Generating a map with size", opt.mapSize)
		param.board = gridGenerator(opt.mapSize)
	} else {
		return errors.New("No valid map size or filename option missing")
	}
	if !isSolvable(param.board) {
		fmt.Fprintln(os.Stderr, "Board is not solvable")
		param.unsolvable = true
	}
	if !opt.noIterativeDepth {
		fmt.Fprintln(os.Stderr, "Search Method : IDA*")
		param.maxScore = param.eval.fx(param.board, param.board, goal(len(param.board)), []byte{}) + 1
	} else {
		fmt.Fprintln(os.Stderr, "Search Method : A*")
		param.maxScore |= (1<<31 - 1)
	}
	return nil
}

func handleSignals() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-sigc
		fmt.Fprintln(os.Stderr, "\b\bExiting after receiving a signal")
		os.Exit(1)
	}()
}

func getAvailableRAM() (uint64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("Error while getting info about memory: %v", err)
	}
	availableRAM := v.Free
	return availableRAM, nil
}

func iterateAlgo(param algoParameters, data *safeData) {
	var wg sync.WaitGroup
Iteration:
	for param.maxScore < 1<<31 {
		if param.maxScore != 1<<31-1 {
			fmt.Fprintln(os.Stderr, "cut off is now :", param.maxScore)
		}
		for i := 0; i < param.workers; i++ {
			wg.Add(1)
			go func(param algoParameters, data *safeData, i int) {

				algo(param, data, i)
				wg.Done()
			}(param, data, i)
		}
		wg.Wait()
		switch {
		case data.win == true:
			fmt.Fprintln(os.Stderr, "Found a solution")
			break Iteration
		case data.ramFailure == true:
			fmt.Fprintln(os.Stderr, "RAM Failure")
			break Iteration
		default:
			*data = initData(param)
			param.maxScore += 2
		}
	}
}

func parseFlags(opt *option) {
	flag.StringVar(&opt.filename, "f", "", "usage : -f [filename]")
	flag.IntVar(&opt.mapSize, "s", 3, "usage : -s [board_size]")
	flag.StringVar(&opt.heuristic, "h", "astar_manhattan", "usage : -h [heuristic]")
	flag.IntVar(&opt.workers, "w", 1, "usage : -w [workers] between 1 and 16")
	flag.IntVar(&opt.seenNodesSplit, "split", 1, "usage : -split [setNodesSplit] between 1 and 256")
	flag.IntVar(&opt.speedDisplay, "speed", 100, "usage : -speed [speedDisplay] between 1 and 2048")
	flag.BoolVar(&opt.noIterativeDepth, "no-i", false, "usage : -no-i. Use A* instead of Iterative Depth A* (aka IDA*). Faster but increase memory consumption")
	flag.BoolVar(&opt.debug, "d", false, "usage : -d. Activate debug info")
	flag.BoolVar(&opt.disableUI, "no-ui", false, "usage : -no-ui. Disable pretty display of solution")
	flag.Parse()
}

func initOptionForApiUse(opt *option) {
	opt.filename = "/dev/stdin"
	opt.disableUI = true
	opt.heuristic = "astar_manhattan"
	opt.noIterativeDepth = true
	opt.workers = 4
	opt.seenNodesSplit = 16
}

func solve(cli bool) (result []string) {
	param := algoParameters{}
	opt := &option{}
	if cli {
		parseFlags(opt)
	} else {
		initOptionForApiUse(opt)
	}
	if err := areFlagsOk(opt); err != nil {
		return []string{"FLAGS", err.Error()}
	}
	if err := setParam(opt, &param); err != nil {
		return []string{"PARAM", err.Error()}
	}
	if param.unsolvable {
		fmt.Fprintln(os.Stderr, "Board is unsolvable", param.board)
		return []string{"UNSOLVABLE"}
	}
	fmt.Fprintf(os.Stderr, "Board is : %v\nNow starting with : %v\n", param.board, param.eval.name)
	data := initData(param)
	start := time.Now()
	iterateAlgo(param, &data)
	end := time.Now()
	elapsed := end.Sub(start)
	if data.path != nil {
		for _, value := range data.seenNodes {
			data.closedSetComplexity += len(value)
		}
		fmt.Fprintln(os.Stderr, "Succes with :", param.eval.name, "in ", elapsed.String(), "!")
		fmt.Fprintf(os.Stderr, "len of solution : %v, time complexity / tries : %d, space complexity : %d, score : %d\n", len(data.path), data.tries, data.closedSetComplexity, data.winScore)
		if !opt.disableUI {
			displayBoard(param.board, data.path, param.eval.name, elapsed.String(), data.tries, data.closedSetComplexity, param.workers, param.seenNodesSplit, opt.speedDisplay)
		}
		return []string{"OK", string(data.path)}
	} else if data.ramFailure {
		return []string{"RAM"}
	}
	return []string{"END"}
}

func main() {
	handleSignals()
	cmdline := true
	fmt.Println(solve(cmdline))
}
