package controller

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/fleblay/42-npuzzle/algo"
	"github.com/fleblay/42-npuzzle/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SolveRequest struct {
	Size            int    `json:"size"`
	Board           string `json:"board"`
	PreviousCompute bool   `json:"previousCompute"`
	Disposition     string `json:"disposition"`
}

type Repository struct {
	DB   *gorm.DB
	Algo string
	Jobs *[]string
}

func (repo *Repository) addStringInputToJobs(input string) (err error) {
	if index := algo.Index(*repo.Jobs, input); index != -1 {
		return errors.New("Grid already beeing processed")
	} else {
		*repo.Jobs = append(*repo.Jobs, input)
		return nil
	}
}

func (repo *Repository) removeStringInputFromJobs(input string) (err error) {
	if index := algo.Index(*repo.Jobs, input); index == -1 {
		return errors.New("Grid already removed from jobs")
	} else {
		*repo.Jobs = append((*repo.Jobs)[:index], ((*repo.Jobs)[index+1:])...)
		return nil
	}
}

func GetSolutionByStringInput(solution *models.Solution, db *gorm.DB, stringInput string, disposition string) error {
	scanner := bufio.NewScanner(strings.NewReader(stringInput))
	board, err := algo.ParseInput(scanner)
	if err != nil {
		return err
	}
	hash := algo.MatrixToStringHashOnly(board, ".")
	return solution.GetSolutionByHash(db, hash, disposition)
}

func (repo *Repository) Solve(c *gin.Context) {
	debug.FreeOSMemory()
	opt := &algo.Option{}
	algo.InitOptionForApiUse(opt, repo.Algo)
	var result [3]string
	fallback := false
	solution := &models.Solution{}

	var newRequest SolveRequest
	if err := c.BindJSON(&newRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
		return
	}
	fmt.Fprintln(os.Stderr, "Received request :", newRequest)
	opt.StringInput = strconv.Itoa(newRequest.Size) + " " + newRequest.Board
	if len(*repo.Jobs) > 0 && (repo.Algo == "A*" || repo.Algo == "default") {
		fmt.Fprintln(os.Stderr, "Server already running an A* job")
		c.IndentedJSON(http.StatusOK, gin.H{"status": "BUSY"})
		return
	}
	if err := repo.addStringInputToJobs(opt.StringInput); err != nil {
		fmt.Fprintln(os.Stderr, "Grid already being processed !")
		c.IndentedJSON(http.StatusOK, gin.H{"status": "RUNNING"})
		return
	}
	if err := GetSolutionByStringInput(solution, repo.DB, opt.StringInput, newRequest.Disposition); err == nil && newRequest.PreviousCompute {
		fmt.Fprintln(os.Stderr, "Found entry in DB !")
		c.IndentedJSON(http.StatusOK, gin.H{"status": "DB", "solution": solution.Path, "time": time.Duration(solution.ComputeMs * 1000).String(), "algo": solution.Algo})
		if err := repo.removeStringInputFromJobs(opt.StringInput); err != nil {
			fmt.Fprintln(os.Stderr, "Failure removing grid from running jobs")
		}
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "No entry found in DB (%s) processing request\n", err.Error())
	}
	result, solution = algo.Solve(opt)
	if result[0] == "RAM" && opt.NoIterativeDepth == true && repo.Algo == "default" {
		fmt.Fprintln(os.Stderr, "Solver killed because of RAM, trying again with IDA*")
		repo.Algo = "fallback_IDA"
		opt.NoIterativeDepth = false
		fallback = true
		result, solution = algo.Solve(opt)
	} else if result[0] == "OK" {
		if err := solution.UpdateOrCreateSolution(repo.DB); err != nil {
			fmt.Fprintln(os.Stderr, "Failure to save new solution to DB")
		}
	} else if result[0] == "PARAM" || result[0] == "FLAGS" {
		fmt.Fprintln(os.Stderr, "Wrong parameters or flags for solver init")
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"status":   result[0],
		"solution": result[1],
		"time":     result[2],
		"algo":     repo.Algo,
		"fallback": fallback,
		"workers":  opt.Workers,
	})
	if repo.removeStringInputFromJobs(opt.StringInput) != nil {
		fmt.Fprintln(os.Stderr, "Failure removing grid from running jobs")
	}
	debug.FreeOSMemory()
}

func (repo *Repository) Generate(c *gin.Context) {
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
	}
	board := algo.MatrixToStringHashOnly(algo.GridGenerator(size), " ")
	c.IndentedJSON(http.StatusOK, gin.H{"size": size, "board": board})
}

func (repo *Repository) GetRandomFromDB(c *gin.Context) {
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
	}
	solution := &models.Solution{}
	count, err := solution.GetCountBySize(repo.DB, size)
	fmt.Fprintln(os.Stderr, "Picking random grid from", count, "suitable entries")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Error Counting grids : " + err.Error()})
	}
	err = solution.GetRandomSolutionBySize(repo.DB, size)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Error Retrieving grids : " + err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"size": solution.Size, "board": strings.Join(strings.Split(solution.Hash, "."), " ")})
}

// TODO : Change Solve route in order to use this code for DRY purpose
func (repo *Repository) GetSolution(c *gin.Context) {
	solution := &models.Solution{}
	var newRequest SolveRequest
	if err := c.BindJSON(&newRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
	}
	fmt.Fprintln(os.Stderr, "Received request :", newRequest)
	StringInput := strconv.Itoa(newRequest.Size) + " " + newRequest.Board
	if err := GetSolutionByStringInput(solution, repo.DB, StringInput, newRequest.Disposition); err == nil {
		fmt.Fprintln(os.Stderr, "Found entry in DB !")
		c.IndentedJSON(http.StatusOK, gin.H{"status": "DB", "solution": solution.Path})
		return
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"status": "NOTFOUND"})
	}
}
