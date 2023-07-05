package main

import ()

func matrixToTable(matrix [][]int) []int {
	boardSize := len(matrix)
	table := make([]int, boardSize*boardSize)
	startLine, endLine := 0, boardSize-1
	startColumn, endColumn := 0, boardSize-1
	index := 0
	for startLine <= endLine && startColumn <= endColumn {
		for i := startColumn; i <= endColumn; i++ {
			table[index] = matrix[startLine][i]
			index++
		}
		startLine++
		for i := startLine; i <= endLine; i++ {
			table[index] = matrix[i][endColumn]
			index++
		}
		endColumn--
		if startLine <= endLine {
			for i := endColumn; i >= startColumn; i-- {
				table[index] = matrix[endLine][i]
				index++
			}
			endLine--
		}
		if startColumn <= endColumn {
			for i := endLine; i >= startLine; i-- {
				table[index] = matrix[i][startColumn]
				index++
			}
			startColumn++
		}
	}
	return table
}

func isSolvable(board [][]int) bool {

	board1d := matrixToTable(board)
	inversions := 0

	for i := 0; i < len(board1d); i++ {
		for j := i + 1; j < len(board1d); j++ {
			if board1d[i] > board1d[j] && board1d[i] != 0 && board1d[j] != 0 {
				inversions++
			}
		}
	}
	return inversions%2 == 0
}
