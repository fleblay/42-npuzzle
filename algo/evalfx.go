package algo

import "math"

func dijkstra(pos, startPos, goalPos [][]int, path []byte) int {
	score := len(path) + 1
	return score
}

func greedy_manhattan(pos, startPos, goalPos [][]int, path []byte) int {
	score := 0
	for j, row := range goalPos {
		for i, value := range row {
			if pos[j][i] != value {
				wrongPositon := getValuePostion(pos, value)
				score += int(math.Abs(float64(wrongPositon.X-i)) + math.Abs(float64(wrongPositon.Y-j)))
			}
		}
	}
	return score
}

func greedy_hamming(pos, startPos, goalPos [][]int, path []byte) int {
	score := 0
	for i, row := range goalPos {
		for j, value := range row {
			if pos[i][j] != value {
				score++
			}
		}
	}
	return score
}

func greedy_inv(pos, startPos, goalPos [][]int, path []byte) int {
	flattenedPos := matrixToTableSnail(pos)
	inversion := 0
	for i := range flattenedPos {
		for j := i + 1; j < len(flattenedPos); j++ {
			if flattenedPos[i] > 0 && flattenedPos[j] > 0 && flattenedPos[i] > flattenedPos[j] {
				inversion++
			}
		}
	}
	return inversion
}
func astar_hamming(pos, startPos, goalPos [][]int, path []byte) int {
	return len(path) + 1 + greedy_hamming(pos, startPos, goalPos, path)
}

func astar_manhattan_generator(weight float64) EvalFx {
	return func(pos, startPos, goalPos [][]int, path []byte) int {
		initDist := len(path) + 1
		return initDist + int(weight*float64(greedy_manhattan(pos, startPos, goalPos, path)))
	}
}

func astar_inv(pos, startPos, goalPos [][]int, path []byte) int {
	return len(path) + 1 + greedy_inv(pos, startPos, goalPos, path)
}
