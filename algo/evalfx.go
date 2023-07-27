package algo

import "math"

func dijkstra(pos, startPos, goalPos [][]int, path []byte) int {
	score := len(path) + 1
	return score
}

func greedy_manhattan_conflict(pos, startPos, goalPos [][]int, path []byte) int {
	score := 0
	for j, row := range pos {
		for i, value := range row {
			if value != goalPos[j][i] {
				goodPosition := getValuePostion(goalPos, value)
				score += int(math.Abs(float64(goodPosition.X-i)) + math.Abs(float64(goodPosition.Y-j)))
				if goodPosition.Y != j/* && goodPosition.X != i*/{
					continue
				}
				for k := i + 1; k < len(row); k++ {
					if rightValue, rightValueGoodPos := row[k], getValuePostion(goalPos, row[k]); rightValue != 0 && rightValue != goalPos[j][k] && rightValueGoodPos.Y == j && rightValueGoodPos.X <= i {
						score += 2
					}
				}
				/*
				for k := j + 1; k < len(pos); k++ {
					if downValue, downValueGoodPos := pos[k][i], getValuePostion(goalPos, pos[k][i]); downValue != 0 && downValue != goalPos[k][i] && downValueGoodPos.X == i && downValueGoodPos.Y <= j {
						score += 2
					}
				}
				*/
			}
		}
	}
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

func astar_hamming(pos, startPos, goalPos [][]int, path []byte) int {
	return len(path) + 1 + greedy_hamming(pos, startPos, goalPos, path)
}

func astar_manhattan_generator(weight float64) EvalFx {
	return func(pos, startPos, goalPos [][]int, path []byte) int {
		initDist := len(path) + 1
		return initDist + int(weight*float64(greedy_manhattan(pos, startPos, goalPos, path)))
	}
}

func astar_manhattan_generator_conflict(weight float64) EvalFx {
	return func(pos, startPos, goalPos [][]int, path []byte) int {
		initDist := len(path) + 1
		return initDist + int(weight*float64(greedy_manhattan_conflict(pos, startPos, goalPos, path)))
	}
}
