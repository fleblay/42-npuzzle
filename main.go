package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fleblay/42-npuzzle/algo"
	"github.com/fleblay/42-npuzzle/database"

	"github.com/gin-gonic/gin"
)

func handleFatalError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error :", err.Error())
		os.Exit(1)
	}
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

func parseFlags(opt *algo.Option) {
	flagSet := &flag.FlagSet{}
	flagSet.SetOutput(os.Stderr)

	flagSet.StringVar(&opt.Filename, "f", "", "usage : -f [filename]")
	flagSet.StringVar(&opt.StringInput, "string", "", "usage : -string [input as a string]. Ex : '3 1 2 3 4 5 6 8 7 0'")
	flagSet.IntVar(&opt.MapSize, "s", 3, "usage : -s [board_size]")
	flagSet.StringVar(&opt.Heuristic, "h", "astar_manhattan", "usage : -h [heuristic]")
	flagSet.IntVar(&opt.Workers, "w", 1, "usage : -w [workers] between 1 and 16")
	flagSet.IntVar(&opt.SeenNodesSplit, "split", 1, "usage : -split [setNodesSplit] between 1 and 256")
	flagSet.IntVar(&opt.SpeedDisplay, "speed", 100, "usage : -speed [speedDisplay] between 1 and 2048")
	flagSet.BoolVar(&opt.NoIterativeDepth, "no-i", false, "usage : -no-i. Use A* instead of Iterative Depth A* (aka IDA*). WAY faster but increase A LOT memory consumption")
	flagSet.BoolVar(&opt.Debug, "d", false, "usage : -d. Activate debug info")
	flagSet.BoolVar(&opt.DisableUI, "no-ui", false, "usage : -no-ui. Disable pretty display of solution")

	flagSet.Parse(os.Args[1:])
}

func initOptionForApiUse(opt *algo.Option) {
	opt.DisableUI = true
	opt.Heuristic = "astar_manhattan"
	opt.NoIterativeDepth = true
	opt.Workers = 4
	opt.SeenNodesSplit = 16
	// to remove
	opt.Debug = true
}

type solveRequest struct {
	Size  int    `json:"size"`
	Board string `json:"board"`
}

func main() {
	handleSignals()
	opt := &algo.Option{}
	db, err := database.ConnectDB("solutions.db")
	handleFatalError(err)
	database.CreateModel(db)

	if os.Getenv("API") != "" {

		initOptionForApiUse(opt)
		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()

		router.GET("/", func(c *gin.Context) {
			c.IndentedJSON(http.StatusOK, gin.H{"msg": "Hello world"})
		})

		router.POST("/", func(c *gin.Context) {
			var newRequest solveRequest
			if err := c.BindJSON(&newRequest); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
			}
			fmt.Println(newRequest)
			opt.StringInput = strconv.Itoa(newRequest.Size) + " " + newRequest.Board
			result := algo.Solve(opt, db)
			if result[0] == "RAM" && opt.NoIterativeDepth == true {
				fmt.Fprintln(os.Stderr, "Killed because of RAM, trying again with IDA*")
				opt.NoIterativeDepth = false
				result = algo.Solve(opt, db)
				opt.NoIterativeDepth = true
			}
			if len(result) > 1 {
				c.IndentedJSON(http.StatusOK, gin.H{"status": result[0], "solution": result[1]})
			} else {
				c.IndentedJSON(http.StatusOK, gin.H{"status": result[0]})
			}
		})

		fmt.Println("Now reading request on localhost:8081")
		router.Run("localhost:8081")
	} else {
		parseFlags(opt)
		fmt.Println(algo.Solve(opt, db))
	}
}
