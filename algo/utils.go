package algo

import (
	"io/fs"
	"io/ioutil"
	"log"
	"strconv"
)

func isEqual[T comparable](a, b [][]T) bool {
	for i := range a {
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func isEqualTable[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Index[T comparable](slice []T, toFind T) int {
	for i, v := range slice {
		if v == toFind {
			return i
		}
	}
	return -1
}

func DeepSliceCopyAndAdd[T any](slice []T, elems ...T) []T {
	newSlice := make([]T, len(slice), len(slice)+len(elems))
	copy(newSlice, slice)
	newSlice = append(newSlice, elems...)
	return newSlice
}

func Deep2DSliceCopy[T any](slice [][]T) [][]T {
	newSlice := make([][]T, len(slice))
	for i, row := range slice {
		newSlice[i] = make([]T, len(row))
		for j, value := range row {
			newSlice[i][j] = value
		}
	}
	return newSlice
}

func Max(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func Min(a, b int) int {
	if a >= b {
		return b
	} else {
		return a
	}
}

func Abs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}

func openDir(dir string) []fs.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func MatrixToStringHashOnly(matrix [][]int, separator string) string {

	results := ""

	for i := 0; i < len(matrix); i++ {

		for j := 0; j < len(matrix[i]); j++ {
			results += strconv.Itoa(matrix[i][j]) + separator
		}
	}
	return results
}

/*
func MatrixToStringSelector(matrix [][]int, worker int, seenNodeMap int) (key string, queueIndex int, seenNodeIndex int) {
	if len(matrix) < 10 {
		return matrixToStringOptimal(matrix, worker, seenNodeMap)
	} else {
		return MatrixToStringNoOpti(matrix, worker, seenNodeMap)
	}
}
*/

func MatrixToStringSelector(matrix [][]int, worker int, seenNodeMap int) (key uint64, queueIndex int, seenNodeIndex int) {
		return matrixToUint64(matrix, worker, seenNodeMap)
}

//TODO fx to change for 3x3 or 4x4
func matrixToUint64(matrix [][]int, worker int, seenNodeMap int) (key uint64, queueIndex int, seenNodeIndex int) {

	size := len(matrix)

	spot := 0
	for i := 0; i < size; i++ {

		for j := 0; j < size; j++ {
			queueIndex += matrix[i][j] * (i + 0) * (j + 0)
			seenNodeIndex += matrix[i][j] * (i + 0) * (j + 0)
			spot += 3
			key |= uint64(matrix[i][j])
			if i != 3 || j != 3 {
				key <<= 4
			}
		}
	}
	queueIndex %= worker
	seenNodeIndex %= seenNodeMap
	return key, queueIndex, seenNodeIndex
}

func matrixToStringOptimal(matrix [][]int, worker int, seenNodeMap int) (key string, queueIndex int, seenNodeIndex int) {

	results := make([]byte, len(matrix)*len(matrix)*4)
	size := len(matrix)

	spot := 0
	for i := 0; i < size; i++ {

		for j := 0; j < size; j++ {
			queueIndex += matrix[i][j] * (i + 0) * (j + 0)
			seenNodeIndex += matrix[i][j] * (i + 0) * (j + 0)
			results[spot] = byte(matrix[i][j] / 10)
			results[spot+1] = byte(matrix[i][j] % 10)
			results[spot+2] = '.'
			spot += 3
		}
	}
	queueIndex %= worker
	seenNodeIndex %= seenNodeMap
	return string(results), queueIndex, seenNodeIndex
}

func MatrixToStringNoOpti(matrix [][]int, worker int, seenNodeMap int) (key string, queueIndex int, seenNodeIndex int) {

	results := ""
	size := len(matrix)

	for i := 0; i < size; i++ {

		for j := 0; j < size; j++ {
			queueIndex *= matrix[i][j] * (i + 0) * (j + 0)
			seenNodeIndex *= matrix[i][j] * (i + 0) * (j + 0)
			results += strconv.Itoa(matrix[i][j]) + "."

		}
	}
	queueIndex %= worker
	seenNodeIndex %= seenNodeMap
	return results, queueIndex, seenNodeIndex
}
