package controller

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/fleblay/42-npuzzle/algo"
	"github.com/fleblay/42-npuzzle/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SolveRequest struct {
	Size  int    `json:"size"`
	Board string `json:"board"`
}

type Repository struct {
	DB *gorm.DB
	Algo string
}

func GetSolutionByStringInput(solution *models.Solution, db *gorm.DB, stringInput string) error {
	scanner := bufio.NewScanner(strings.NewReader(stringInput))
	board, err := algo.ParseInput(scanner)
	if err != nil {
		return err
	}
	hash := algo.MatrixToStringHashOnly(board, ".")
	return solution.GetSolutionByHash(db, hash)
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
	}
	fmt.Fprintln(os.Stderr, "Received request :", newRequest)
	opt.StringInput = strconv.Itoa(newRequest.Size) + " " + newRequest.Board

	if err := GetSolutionByStringInput(solution, repo.DB, opt.StringInput); err == nil {
		fmt.Fprintln(os.Stderr, "Found entry in DB !")
		c.IndentedJSON(http.StatusOK, gin.H{"status": "DB", "solution": solution.Path})
		return
	} else {
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
	} else if result[0] == "PARAM" || result[0] == "FLAGS"{
			fmt.Fprintln(os.Stderr, "Wrong parameters or flags for solver init")
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"status": result[0],
		"solution": result[1],
		"time": result[2],
		"algo" : repo.Algo,
		"fallback" : fallback,
		"workers" : opt.Workers,
	})
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

//TODO : Change Solve route in order to use this code for DRY purpose
func (repo *Repository) GetSolution(c *gin.Context) {
	solution := &models.Solution{}
	var newRequest SolveRequest
	if err := c.BindJSON(&newRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
	}
	fmt.Fprintln(os.Stderr, "Received request :", newRequest)
	StringInput := strconv.Itoa(newRequest.Size) + " " + newRequest.Board
	if err := GetSolutionByStringInput(solution, repo.DB, StringInput); err == nil {
		fmt.Fprintln(os.Stderr, "Found entry in DB !")
		c.IndentedJSON(http.StatusOK, gin.H{"status": "DB", "solution": solution.Path})
		return
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"status" : "NOTFOUND"})
	}
}
