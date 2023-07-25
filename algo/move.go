package algo

func getValuePostion(board [][]int, toFind int) Pos2D {
	for i, line := range board {
		for j, value := range line {
			if value == toFind {
				return Pos2D{X: j, Y: i}
			}
		}
	}
	return Pos2D{-1, -1}
}

func swap[T any](a, b *T) {
	*a, *b = *b, *a
}

func moveUp(board [][] int) (ok bool, updatedBoard [][]int) {
	empty := getValuePostion(board, 0)
	updatedBoard = Deep2DSliceCopy(board)
	if empty.Y != 0 {
		swap(&updatedBoard[empty.Y][empty.X], &updatedBoard[empty.Y-1][empty.X])
		return true, updatedBoard
	}
	return false, nil
}

func moveDown(board [][]int) (ok bool, updatedBoard [][]int) {
	empty := getValuePostion(board, 0)
	updatedBoard = Deep2DSliceCopy(board)
	if empty.Y != len(board)-1 {
		swap(&updatedBoard[empty.Y][empty.X], &updatedBoard[empty.Y+1][empty.X])
		return true, updatedBoard
	}
	return false, nil
}

func moveLeft(board [][]int) (ok bool, updatedBoard [][]int) {
	empty := getValuePostion(board, 0)
	updatedBoard = Deep2DSliceCopy(board)
	if empty.X != 0 {
		swap(&updatedBoard[empty.Y][empty.X], &updatedBoard[empty.Y][empty.X-1])
		return true, updatedBoard
	}
	return false, nil
}

func moveRight(board [][]int) (ok bool, updatedBoard [][]int) {
	empty := getValuePostion(board, 0)
	updatedBoard = Deep2DSliceCopy(board)
	if empty.X != len(board)-1 {
		swap(&updatedBoard[empty.Y][empty.X], &updatedBoard[empty.Y][empty.X+1])
		return true, updatedBoard
	}
	return false, nil
}
