package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/fleblay/42-npuzzle/algo"
	"github.com/fleblay/42-npuzzle/controller"
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
	flagSet.StringVar(&opt.StringInput, "string", "", "usage : -string [input as a string, starting with the size]. Ex : '3 1 2 3 4 5 6 8 7 0'")
	flagSet.IntVar(&opt.MapSize, "s", 3, "usage : -s [board_size]. Use a board randomly generated of selected size")
	flagSet.StringVar(&opt.Heuristic, "h", "astar_manhattan", "usage : -h [heuristic]")
	flagSet.IntVar(&opt.Workers, "w", 1, "usage : -w [workers] between 1 and 16")
	flagSet.IntVar(&opt.SeenNodesSplit, "split", 1, "usage : -split [setNodesSplit] between 1 and 256")
	flagSet.IntVar(&opt.SpeedDisplay, "speed", 100, "usage : -speed [speedDisplay] between 1 and 2048")
	flagSet.BoolVar(&opt.NoIterativeDepth, "no-i", false, "usage : -no-i. Use A* instead of Iterative Depth A* (aka IDA*). WAY faster but increase A LOT memory consumption")
	flagSet.BoolVar(&opt.Debug, "d", false, "usage : -d. Activate debug info")
	flagSet.BoolVar(&opt.DisableUI, "no-ui", false, "usage : -no-ui. Disable pretty display of solution")

	flagSet.Parse(os.Args[1:])
}

func main() {
	handleSignals()

	if os.Getenv("API") == "true" {
		db, err := database.ConnectDB("solutions.db")
		handleFatalError(err)
		count, err := database.CreateModel(db)
		fmt.Printf("Successfully connected to DB with %d items\n", count)
		handleFatalError(err)
		repo := controller.Repository{DB: db}

		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()

		router.POST("/solve", repo.Solve)
		router.POST("/solution", repo.GetSolution)
		router.GET("/generate/:size", repo.Generate)
		router.GET("/pick/:size", repo.GetRandomFromDB)

		listen := os.Getenv("LISTEN")
		if listen != "" {
			fmt.Printf("Starting server on '%s'\n", listen)
			err = router.Run(listen)
		} else {
			fmt.Println("Starting server with default value 'localhost:8081'")
			err = router.Run("localhost:8081")
		}
		handleFatalError(err)
	} else {
		opt := &algo.Option{}
		parseFlags(opt)
		res, _ := algo.Solve(opt)
		fmt.Println(res)
	}
}
