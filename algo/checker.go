package algo

import (
	"fmt"
	"os"
)

func matrixToTableSnail(matrix [][]int) []int {
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

func flattenBoard(matrix [][]int) (flatBoard []int) {
	boardSize := len(matrix)
	flatBoard = make([]int, boardSize*boardSize)
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			flatBoard[i*boardSize+j] = matrix[i][j]
		}
	}
	return
}

func isSolvableZeroLast(board [][]int) (ok bool, inversions int) {
	odd := (len(board) % 2) != 0
	oddRowCountFromBottomToZero := ((len(board)-1)-getValuePostion(board, 0).Y)%2 == 0
	board1d := flattenBoard(board)

	for i := 0; i < len(board1d); i++ {
		for j := i + 1; j < len(board1d); j++ {
			if board1d[i] > board1d[j] && board1d[i] != 0 && board1d[j] != 0 {
				inversions++
			}
		}
	}
	if odd {
		fmt.Fprintln(os.Stderr, "grid size is odd and inversion count is", inversions)
		return inversions%2 == 0, inversions
	} else {
		fmt.Fprintln(os.Stderr, "grid size is odd and inversion count is", inversions, "and pos 0 from bottom is odd ? : ", oddRowCountFromBottomToZero)
		if (!oddRowCountFromBottomToZero && (inversions%2 == 1)) ||
			(oddRowCountFromBottomToZero && (inversions%2 == 0)) {
			return true, inversions
		}
		return false, inversions
	}
}

func isSolvableSnail(board [][]int) (ok bool, inversions int) {
	board1d := matrixToTableSnail(board)

	for i := 0; i < len(board1d); i++ {
		for j := i + 1; j < len(board1d); j++ {
			if board1d[i] > board1d[j] && board1d[i] != 0 && board1d[j] != 0 {
				inversions++
			}
		}
	}
	return inversions%2 == 0, inversions
}

func IsSolvable(board [][]int, disposition string) (ok bool, inversions int) {
	if disposition == "snail" {
		return isSolvableSnail(board)
	}
	return isSolvableZeroLast(board)
}
