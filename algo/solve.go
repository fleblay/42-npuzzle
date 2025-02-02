package algo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/fleblay/42-npuzzle/models"
)

func InitOptionForApiUse(opt *Option, algo string) {
	opt.DisableUI = true
	opt.Heuristic = "astar_manhattan_conflict"
	if algo == "A*" {
		opt.NoIterativeDepth = true
	}
	opt.Workers = 8
	opt.SeenNodesSplit = 96
	opt.RAMMaxGB = 6
	//To be changed in prod
	opt.Debug = true
}

func areFlagsOk(opt *Option) (err error) {
	if opt.Workers < 1 || opt.Workers > 32 {
		return errors.New("Invalid number of workers")
	}
	if opt.SeenNodesSplit < 1 || opt.SeenNodesSplit > 96 {
		return errors.New("Invalid number of splits")
	}
	if opt.Filename == "" && opt.StringInput == "" && (opt.MapSize < 3 || opt.MapSize > 4) {
		return errors.New("Invalid map size")
	}
	if opt.RAMMaxGB < 1 || opt.RAMMaxGB > 64 {
		return errors.New("Invalid Max Ram GB (must be between 1 and 32GB")
	}
	if opt.Disposition != "snail" && opt.Disposition != "zerolast" {
		return errors.New("Invalid disposition")
	}
	for _, current := range Evals {
		if current.Name == opt.Heuristic {
			return nil
		}
	}
	return errors.New("Invalid heuristic")
}

func setParam(opt *Option, param *AlgoParameters) (err error) {
	param.Workers = opt.Workers
	param.SeenNodesSplit = opt.SeenNodesSplit
	param.Disposition = opt.Disposition
	for _, current := range Evals {
		if current.Name == opt.Heuristic {
			param.Eval = current
			break
		}
	}
	if !opt.Debug {
		os.Stderr.Close()
		newstderr, err := os.Open(os.DevNull)
		if err != nil {
			return err
		}
		os.Stderr = newstderr
	}
	if opt.Filename != "" {
		fmt.Fprintln(os.Stderr, "Opening user provided map in file", opt.Filename)
		opt.Fd, err = OpenFile(opt.Filename)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(opt.Fd)
		param.Board, err = ParseInput(scanner)
		opt.Fd.Close()
	} else if opt.StringInput != "" {
		fmt.Fprintln(os.Stderr, "Reading from provided string", opt.StringInput)
		scanner := bufio.NewScanner(strings.NewReader(opt.StringInput))
		param.Board, err = ParseInput(scanner)
	} else if opt.MapSize > 0 {
		fmt.Fprintln(os.Stderr, "Generating a map with size", opt.MapSize)
		param.Board = GridGenerator(opt.MapSize, param.Disposition)
	} else {
		return errors.New("No valid filename, stringMap or mapSize")
	}
	if err != nil {
		return err
	}
	if ok, _ := IsSolvable(param.Board, param.Disposition); !ok {
		fmt.Fprintln(os.Stderr, "Board is not solvable")
		param.Unsolvable = true
		return errors.New("Board is not solvable")
	}
	param.RAMMaxGB = opt.RAMMaxGB
	if opt.RAMMaxGB > 1 && opt.NoIterativeDepth {
		fmt.Fprintf(os.Stderr, "Solver will use a soft maxmimum of %d Gb and a hard maximum of %d Gb of RAM\n", param.RAMMaxGB-1, param.RAMMaxGB)
		debug.SetMemoryLimit(int64((opt.RAMMaxGB - 1) << 30))
		debug.SetGCPercent(-1)
	} else {
		fmt.Fprintf(os.Stderr, "Solver will use a soft maxmimum of %d Gb. RAM failure will be triggered if available RAM drops below %d Mb\n", param.RAMMaxGB, MinRAMAvailableMB)
		debug.SetMemoryLimit(int64(opt.RAMMaxGB << 30))
		debug.SetGCPercent(200)
	}
	return err
}

func displayResult(algoResult Result, opt Option, param AlgoParameters, elapsed time.Duration) {
	fmt.Fprintln(os.Stderr, "Succes with :", param.Eval.Name, "in ", elapsed.String(), "!")
	fmt.Fprintf(os.Stderr, "len of solution : %v, time complexity / tries : %d, space complexity : %d\n", len(algoResult.Path), algoResult.Tries, algoResult.ClosedSetComplexity)
	if !opt.DisableUI {
		DisplayBoard(param.Board, algoResult.Path, param.Eval.Name, elapsed.String(), algoResult.Tries, algoResult.ClosedSetComplexity, param.Workers, param.SeenNodesSplit, opt.SpeedDisplay)
	}
}

func generateSolutionEntity(param AlgoParameters, algoResult Result, elapsed time.Duration) *models.Solution {
	solution := models.Solution{
		Size:        len(param.Board),
		Hash:        MatrixToStringHashOnly(param.Board, "."),
		Path:        string(algoResult.Path),
		Length:      len(algoResult.Path),
		Algo:        algoResult.Algo,
		Solvable:    true,
		Workers:     param.Workers,
		Split:       param.SeenNodesSplit,
		Disposition: param.Disposition,
		ComputeMs:   elapsed.Microseconds(),
	}
	return &solution
}

func Solve(opt *Option) (result [3]string, solution *models.Solution) {
	param := AlgoParameters{}
	algoResult := Result{}
	if err := areFlagsOk(opt); err != nil {
		return [3]string{"FLAGS", err.Error()}, nil
	}
	if err := setParam(opt, &param); err != nil {
		return [3]string{"PARAM", err.Error()}, nil
	}
	fmt.Fprintf(os.Stderr, "Board is : %v\nNow starting with : %v\n", param.Board, param.Eval.Name)
	start := time.Now()
	if opt.NoIterativeDepth {
		data := initData(param)
		algoResult = launchAstarWorkers(param, &data)
	} else {
		data := initDataIDA(param)
		algoResult = iterateIDA(&data)
	}
	elapsed := time.Now().Sub(start)
	if algoResult.Path != nil {
		displayResult(algoResult, *opt, param, elapsed)
		return [3]string{"OK", string(algoResult.Path), elapsed.String()}, generateSolutionEntity(param, algoResult, elapsed)
	} else if algoResult.RamFailure {
		return [3]string{"RAM", strconv.Itoa(algoResult.ClosedSetComplexity), elapsed.String()}, nil
	}
	return [3]string{"END"}, nil
}
