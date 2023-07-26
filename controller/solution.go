package controller

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
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
}

func GetSolutionByStringInput(solution *models.Solution, db *gorm.DB, stringInput string) error {
	scanner := bufio.NewScanner(strings.NewReader(stringInput))
	board, err := algo.ParseInput(scanner)
	if err != nil {
		return err
	}
	hash := algo.MatrixToStringHashOnly(board)
	return solution.GetSolutionByHash(db, hash)
}

func (repo *Repository) Solve(c *gin.Context) {
	opt := &algo.Option{}
	algo.InitOptionForApiUse(opt)
	var result [2]string
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
	if result[0] == "RAM" && opt.NoIterativeDepth == true {
		fmt.Fprintln(os.Stderr, "Solver killed because of RAM, trying again with IDA*")
		opt.NoIterativeDepth = false
		result, solution = algo.Solve(opt)
	}
	if result[0] == "OK" {
		if err := solution.UpdateOrCreateSolution(repo.DB); err != nil {
			fmt.Fprintln(os.Stderr, "Failure to save new solution to DB")
		}
	}
	c.IndentedJSON(http.StatusOK, gin.H{"status": result[0], "solution": result[1]})
}
