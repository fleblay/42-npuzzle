package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

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

func (repo *Repository) Solve(c *gin.Context) {
	opt := &algo.Option{}
	algo.InitOptionForApiUse(opt)
	var result []string
	var solution *models.Solution

	var newRequest SolveRequest
	if err := c.BindJSON(&newRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"msg": "Wrong Format : " + err.Error()})
	}
	fmt.Println(newRequest)
	opt.StringInput = strconv.Itoa(newRequest.Size) + " " + newRequest.Board
	result, solution = algo.Solve(opt)
	if result[0] == "RAM" && opt.NoIterativeDepth == true {
		fmt.Fprintln(os.Stderr, "Killed because of RAM, trying again with IDA*")
		opt.NoIterativeDepth = false
		result, solution = algo.Solve(opt)
		opt.NoIterativeDepth = true
	}
	if result[0] == "OK" {
		if err := solution.UpdateOrCreateSolution(repo.DB); err != nil {
			fmt.Fprintln(os.Stderr, "Failure to save new solution to DB")
		}
	}
	// To change
	if len(result) > 1 {
		c.IndentedJSON(http.StatusOK, gin.H{"status": result[0], "solution": result[1]})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"status": result[0]})
	}
	// To change
}
