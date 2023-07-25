package algo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Board struct {
	Board [][]int
}

func OpenFile(filename string) (fd *os.File, err error) {
	fd, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func ParseInput(scanner *bufio.Scanner) (board [][]int, err error) {
	scanner.Split(bufio.ScanLines)
	inputArray, err := readInputArray(scanner)
	if err != nil {
		return nil, err
	}
	size, err := extractSize(inputArray)
	if err != nil {
		return nil, err
	}
	board, err = createBoard(size, inputArray)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func alreadyInArray(array []int, num int) bool {
	if len(array) == 0 {
		return false
	}
	for i := 1; i < len(array); i++ {
		if array[i] == num {
			return true
		}
	}
	return false
}

func readInputArray(scanner *bufio.Scanner) (inputArray []int, err error) {
	inputArray = make([]int, 0, 100)

	for scanner.Scan() {
		line := scanner.Text()
		comment := strings.Index(line, "#")
		if comment != -1 {
			line = line[:comment]
		}
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		words := strings.Fields(line)
		for _, word := range words {
			num, err := strconv.Atoi(word)
			if err != nil || num < 0 {
				return nil, errors.New("Error parsing input : Atoi Error or number < 0")
			} else if alreadyInArray(inputArray, num) {
				return nil, errors.New("Error parsing input : duplicate number")
			}
			inputArray = append(inputArray, num)
		}
	}
	return inputArray, nil
}

func extractSize(inputArray []int) (size int, err error) {
	if len(inputArray) == 0 {
		return -1, errors.New("Error parsing input : wrong grid size")
	}
	size = inputArray[0]
	if size < 3 {
		return -1, errors.New("Error parsing input : grid size is below 3")
	}
	if size*size > len(inputArray)-1 {
		return -1, errors.New("Error parsing input : missing numbers in grid")
	}
	if size*size < len(inputArray)-1 {
		return -1, errors.New("Error parsing input : extra numbers in grid")
	}
	return size, nil
}

func createBoard(size int, inputArray []int) (board [][]int, err error) {
	board = make([][]int, size)
	for i := 0; i < size; i++ {
		board[i] = make([]int, size)
		for j := 0; j < size; j++ {
			if inputArray[i*size+j+1] < 0 || inputArray[i*size+j+1] > size*size-1 {
				return nil, errors.New(fmt.Sprintln("Error parsing input. Size :", size, "number causing error : [", strconv.Itoa(inputArray[i*size+j+1]), "] at row :", i, "and column :", j))
			}
			board[i][j] = inputArray[i*size+j+1]
		}
	}
	return board, nil
}
