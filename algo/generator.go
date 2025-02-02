package algo

func GridGenerator(mapSize int, disposition string) (board [][]int) {

	for {
		randomNumber := make(map[int]int)
		for i := 0; i < mapSize*mapSize; i++ {
			randomNumber[i] = i
		}
		board = make([][]int, mapSize)

		j := 0
		i := 0
		for _, number := range randomNumber {
			if i%mapSize == 0 {
				board[j] = make([]int, mapSize)
				j++
			}
			board[j-1][i%mapSize] = number
			i++
		}
		if ok, inversion := IsSolvable(board, disposition); ok && inversion < 12 {
			break
		}
	}
	return board
}

func Goal(mapSize int, disposition string) (goal [][]int) {
	switch disposition {
	case "snail":
		return snailGoal(mapSize)
	case "zerolast":
		return zeroLastGoal(mapSize)
	}
	return
}

func zeroLastGoal(mapSize int) (goal [][]int) {
	goal = make([][]int, mapSize)
	for i := range goal {
		goal[i] = make([]int, mapSize)
		for j := 0; j < mapSize; j++ {
			if i != mapSize-1 || j != mapSize-1 {
				goal[i][j] = i*mapSize + j + 1
			}
		}
	}
	return
}

func snailGoal(mapSize int) (goal [][]int) {

	goal = make([][]int, mapSize)
	for i := range goal {
		goal[i] = make([]int, mapSize)
	}
	states := []Move2D{
		{'r', 1, 0},
		{'d', 0, 1},
		{'l', -1, 0},
		{'u', 0, -1},
	}
	goal[0][0] = 1
	for i, j, dir, count := 0, 0, 0, 1; count < (mapSize*mapSize)-1; {
		currMove := states[dir%4]
		nextJ := j + currMove.Y
		nextI := i + currMove.X
		if nextI > mapSize-1 ||
			nextI < 0 ||
			nextJ > mapSize-1 ||
			nextJ < 0 ||
			goal[nextJ][nextI] != 0 {
			dir++
		} else {
			i, j = nextI, nextJ
			count++
			goal[j][i] = count
		}
	}
	return goal
}
