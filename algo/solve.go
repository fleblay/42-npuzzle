package algo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fleblay/42-npuzzle/models"
)

func InitOptionForApiUse(opt *Option) {
	opt.DisableUI = true
	opt.Heuristic = "astar_manhattan"
	opt.NoIterativeDepth = true
	opt.Workers = 4
	opt.SeenNodesSplit = 16
	opt.Debug = true
}


func areFlagsOk(opt *Option) (err error) {
	if opt.Workers < 1 || opt.Workers > 32 {
		return errors.New("Invalid number of workers")
	}
	if opt.SeenNodesSplit < 1 || opt.SeenNodesSplit > 4096 {
		return errors.New("Invalid number of splits")
	}
	if opt.Filename == "" && opt.StringInput == "" && (opt.MapSize < 3) {
		return errors.New("Invalid map size")
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
		param.Board = GridGenerator(opt.MapSize)
	} else {
		return errors.New("No valid filename, stringMap or mapSize")
	}
	if err != nil {
		return err
	}
	if ok, _ := IsSolvable(param.Board) ; !ok {
		fmt.Fprintln(os.Stderr, "Board is not solvable")
		param.Unsolvable = true
	}
	return nil
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
		Disposition: "snail",
		ComputeMs:   elapsed.Milliseconds(),
	}
	return &solution
}

func Solve(opt *Option) (result [2]string, solution *models.Solution) {
	param := AlgoParameters{}
	algoResult := Result{}
	if err := areFlagsOk(opt); err != nil {
		return [2]string{"FLAGS", err.Error()}, nil
	}
	if err := setParam(opt, &param); err != nil {
		return [2]string{"PARAM", err.Error()}, nil
	}
	if param.Unsolvable {
		fmt.Fprintln(os.Stderr, "Board is unsolvable", param.Board)
		return [2]string{"UNSOLVABLE"}, nil
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
		return [2]string{"OK", string(algoResult.Path)}, generateSolutionEntity(param, algoResult, elapsed)
	} else if algoResult.RamFailure {
		return [2]string{"RAM"}, nil
	}
	return [2]string{"END"}, nil
}

