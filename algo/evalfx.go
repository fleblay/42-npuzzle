package algo

import (
	//"fmt"
	"math"
)

func dijkstra(pos, startPos, goalPos [][]int, path []byte) int {
	score := len(path) + 1
	return score
}

func Greedy_conflict(pos, startPos, goalPos [][]int, path []byte) int {
	conflit := 0
	for j, row := range pos {
		var conflictLine ConflictGraph
		conflictLine.Init()
		for i, value := range row {
			if value != goalPos[j][i] && value != 0 {
				goodPosition := getValuePostion(goalPos, value)
				if goodPosition.Y == j {
					for k := i + 1; k < len(row); k++ {
						rightValue, rightValueGoodPos := row[k], getValuePostion(goalPos, row[k])
						if rightValue != 0 && rightValue != goalPos[j][k] &&
							rightValueGoodPos.Y == j && rightValueGoodPos.X < goodPosition.X {
							conflictLine.Add(i, k)
							//fmt.Printf("%d : conflict with %d\n", row[i], row[k])
						}
					}
				}
			}
		}
		conf := conflictLine.PopAndCount()
		//fmt.Printf("row[%d] : %d\n", j, conf)
		conflit += conf
	}
	conflictCol := make([]ConflictGraph, len(pos))
	for index := range conflictCol {
		conflictCol[index].Init()
	}
	for j, row := range pos {
		for i, value := range row {
			if value != goalPos[j][i] && value != 0 {
				goodPosition := getValuePostion(goalPos, value)
				if goodPosition.X == i {
					for k := j + 1; k < len(pos); k++ {
						downValue, downValueGoodPos := pos[k][i], getValuePostion(goalPos, pos[k][i])
						if downValue != 0 && downValue != goalPos[k][i] &&
							downValueGoodPos.X == i && downValueGoodPos.Y < goodPosition.Y {
							conflictCol[i].Add(j, k)
						}
					}
				}
			}
		}
	}
	for index := range conflictCol {
		conflit += conflictCol[index].PopAndCount()
	}
	return 2 * conflit
}

func greedy_manhattan(pos, startPos, goalPos [][]int, path []byte) int {
	score := 0
	for j, row := range pos {
		for i, value := range row {
			if value != goalPos[j][i] && value != 0 {
				goodPositon := getValuePostion(goalPos, value)
				score += int(math.Abs(float64(goodPositon.X-i)) + math.Abs(float64(goodPositon.Y-j)))
			}
		}
	}
	return score
}

func greedy_hamming(pos, startPos, goalPos [][]int, path []byte) int {
	score := 0
	for i, row := range goalPos {
		for j, value := range row {
			if pos[i][j] != value && value != 0 {
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
		return initDist + int(weight*(float64(greedy_manhattan(pos, startPos, goalPos, path))+float64(Greedy_conflict(pos, startPos, goalPos, path))))
	}
}
