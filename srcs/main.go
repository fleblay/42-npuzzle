// os.exit a remove
// generator avec une map size a remove
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"github.com/shirou/gopsutil/v3/mem"
)
/*
import (
	"net/http"
	"github.com/gin-gonic/gin"
)
*/


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
	if opt.filename == "" && opt.stringInput == "" && (opt.mapSize < 3 || opt.mapSize > 10) {
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
		scanner := bufio.NewScanner(opt.fd)
		param.board, err = ParseInput(scanner)
		opt.fd.Close()
	} else if opt.stringInput != "" {
		fmt.Fprintln(os.Stderr, "Reading from provided string", opt.stringInput)
		scanner := bufio.NewScanner(strings.NewReader(opt.stringInput))
		param.board, err = ParseInput(scanner)
	} else if opt.mapSize > 0 {
		fmt.Fprintln(os.Stderr, "Generating a map with size", opt.mapSize)
		param.board = gridGenerator(opt.mapSize)
	} else {
		return errors.New("No valid map size or filename option missing")
	}
	if err != nil {
		return err
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
	flagSet := &flag.FlagSet{}
	flagSet.SetOutput(os.Stderr)

	flagSet.StringVar(&opt.filename, "f", "", "usage : -f [filename]")
	flagSet.StringVar(&opt.stringInput, "string", "", "usage : -string [input as a string]. Ex : '3 1 2 3 4 5 6 8 7 0'")
	flagSet.IntVar(&opt.mapSize, "s", 3, "usage : -s [board_size]")
	flagSet.StringVar(&opt.heuristic, "h", "astar_manhattan", "usage : -h [heuristic]")
	flagSet.IntVar(&opt.workers, "w", 1, "usage : -w [workers] between 1 and 16")
	flagSet.IntVar(&opt.seenNodesSplit, "split", 1, "usage : -split [setNodesSplit] between 1 and 256")
	flagSet.IntVar(&opt.speedDisplay, "speed", 100, "usage : -speed [speedDisplay] between 1 and 2048")
	flagSet.BoolVar(&opt.noIterativeDepth, "no-i", false, "usage : -no-i. Use A* instead of Iterative Depth A* (aka IDA*). Faster but increase memory consumption")
	flagSet.BoolVar(&opt.debug, "d", false, "usage : -d. Activate debug info")
	flagSet.BoolVar(&opt.disableUI, "no-ui", false, "usage : -no-ui. Disable pretty display of solution")

	flagSet.Parse(os.Args[1:])
}

func initOptionForApiUse(opt *option) {
	opt.disableUI = true
	opt.heuristic = "astar_manhattan"
	opt.noIterativeDepth = true
	opt.workers = 4
	opt.seenNodesSplit = 16
}

func solve(cli bool, stringInput string) (result []string) {
	param := algoParameters{}
	opt := &option{}
	if cli {
		parseFlags(opt)
	} else {
		initOptionForApiUse(opt)
		opt.stringInput = stringInput
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

type solveRequest struct {
	Size  int `json:"size"`
	Board string `json:"board"`
}

func main() {
	handleSignals()

	/*
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, gin.H{"msg": "Hello world"})
	})

	//curl -X POST --data '{"size":3,"board":"1 2 3 4 5 6 7 8 0"}' localhost:8080
	router.POST("/", func(c *gin.Context) {
		var newRequest solveRequest
		if err := c.BindJSON(&newRequest); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
		}
		fmt.Println(newRequest)
	})

	router.Run("localhost:8080")
	*/
	//fmt.Println(solve(false, "3 1 2 3 4 5 6 8 7 0"))
	fmt.Println(solve(true, ""))
}
