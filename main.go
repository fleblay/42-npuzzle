package main

import (
	"flag"
	"fmt"

	"github.com/fleblay/42-npuzzle/algo"
	"github.com/fleblay/42-npuzzle/controller"
	"github.com/fleblay/42-npuzzle/database"
	"github.com/gin-gonic/gin"

	//FOR CORS
	//cors "github.com/rs/cors/wrapper/gin"

	//FOR PPROF
	//"net/http"
	//"sync"
	//_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
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
	flagSet.StringVar(&opt.Heuristic, "h", "astar_manhattan_conflict", "usage : -h [heuristic]")
	flagSet.IntVar(&opt.Workers, "w", 8, "usage : -w [workers] between 1 and 32")
	flagSet.IntVar(&opt.SeenNodesSplit, "split", 96, "usage : -split [setNodesSplit] between 1 and 96")
	flagSet.IntVar(&opt.SpeedDisplay, "speed", 100, "usage : -speed [speedDisplay] between 1 and 2048")
	flagSet.BoolVar(&opt.NoIterativeDepth, "no-i", false, "usage : -no-i. Use A* instead of Iterative Depth A* (aka IDA*). WAY faster but increase A LOT memory consumption")
	flagSet.BoolVar(&opt.Debug, "d", false, "usage : -d. Activate debug info")
	flagSet.BoolVar(&opt.DisableUI, "no-ui", false, "usage : -no-ui. Disable pretty display of solution")
	flagSet.Uint64Var(&opt.RAMMaxGB, "ram", 8, "usage : -ram [MaxRamGb] between 1 and 16")
	flagSet.StringVar(&opt.Disposition, "dispo", "snail", "usage : -dispo [snail | zerolast]")

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
		repoASTAR := controller.Repository{DB: db, Algo: "A*", Jobs: &[]string{}}
		repoIDA := controller.Repository{DB: db, Algo: "IDA", Jobs: &[]string{}}

		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()

		//Should ONLY be used for testing in dev env
		//router.Use(cors.Default())

		router.POST("/solve/ida", repoIDA.Solve)
		router.POST("/solve/astar", repoASTAR.Solve)
		router.POST("/solution", repoIDA.GetSolution)
		router.GET("/generate/:size/:disposition", repoIDA.Generate)
		router.GET("/pick/:size", repoIDA.GetRandomFromDB)

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
		/*
		var wg sync.WaitGroup
		go func() {
			fmt.Println(http.ListenAndServe("localhost:6060", nil))
			wg.Add(1)
		}()
		*/
		opt := &algo.Option{}
		parseFlags(opt)
		res, _ := algo.Solve(opt)
		fmt.Println(res)
		//wg.Wait()
	}
}
