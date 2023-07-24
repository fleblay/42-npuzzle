package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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
	if opt.filename == "" && opt.stringInput == "" && (opt.mapSize < 3) {
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
	availableRAM := v.Available
	return availableRAM, nil
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
	flagSet.BoolVar(&opt.noIterativeDepth, "no-i", false, "usage : -no-i. Use A* instead of Iterative Depth A* (aka IDA*). WAY faster but increase A LOT memory consumption")
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
	// to remove
	opt.debug = true
}

func displayResult(algoResult Result, opt option, param algoParameters, elapsed time.Duration) {
	fmt.Fprintln(os.Stderr, "Succes with :", param.eval.name, "in ", elapsed.String(), "!")
	fmt.Fprintf(os.Stderr, "len of solution : %v, time complexity / tries : %d, space complexity : %d\n", len(algoResult.path), algoResult.tries, algoResult.closedSetComplexity)
	if !opt.disableUI {
		displayBoard(param.board, algoResult.path, param.eval.name, elapsed.String(), algoResult.tries, algoResult.closedSetComplexity, param.workers, param.seenNodesSplit, opt.speedDisplay)
	}
}

func solve(opt *option) (result []string) {
	param := algoParameters{}
	algoResult := Result{}
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
	start := time.Now()
	if opt.noIterativeDepth {
		data := initData(param)
		algoResult = iterateAlgo(param, &data)
	} else {
		data := initDataIDA(param)
		algoResult = iterateIDA(&data)
	}
	elapsed := time.Now().Sub(start)
	if algoResult.path != nil {
		displayResult(algoResult, *opt, param, elapsed)
		return []string{"OK", string(algoResult.path)}
	} else if algoResult.ramFailure {
		return []string{"RAM"}
	}
	return []string{"END"}
}

type solveRequest struct {
	Size  int    `json:"size"`
	Board string `json:"board"`
}

func main() {
	handleSignals()
	opt := &option{}

	if os.Getenv("API") != "" {
		initOptionForApiUse(opt)
		gin.SetMode(gin.ReleaseMode)
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
			opt.stringInput = strconv.Itoa(newRequest.Size) + " " + newRequest.Board
			result := solve(opt)
			if result[0] == "RAM" && opt.noIterativeDepth == true {
				fmt.Fprintln(os.Stderr, "Killed because of RAM, trying again with IDA*")
				opt.noIterativeDepth = false
				result = solve(opt)
			}
			if len(result) > 1 {
				c.IndentedJSON(http.StatusOK, gin.H{"status": result[0], "solution": result[1]})
			} else {
				c.IndentedJSON(http.StatusOK, gin.H{"status": result[0]})
			}
		})

		fmt.Println("Now reading request on localhost:8080")
		router.Run("localhost:8080")
	} else {
		parseFlags(opt)
		fmt.Println(solve(opt))
	}
}
