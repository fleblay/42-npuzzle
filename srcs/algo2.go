package main

import (
	"fmt"
	"os"
)

func initDataIDA(param algoParameters) (data idaData) {
	data.maxScore = param.eval.fx(param.board, param.board, goal(len(param.board)), []byte{}) + 1
	data.states = append(data.states, Deep2DSliceCopy(param.board))
	hash, _, _ := matrixToStringSelector(param.board, 1, 1)
	data.hashes = append(data.hashes, hash)
	data.fx = param.eval.fx
	data.goal = goal(len(param.board))
	return
}

func iterateIDA(data *idaData) {
	for data.maxScore < 1<<30 {
		fmt.Fprintln(os.Stderr, "true IDA* cut off is now :", data.maxScore)
		newMaxScore, found := IDA(data)
		if found {
			fmt.Fprintln(os.Stderr, "Solution !")
			return
		}
		data.maxScore = newMaxScore
	}
	fmt.Fprintln(os.Stderr, "No solution")
	data.path = nil
}

func IDA(data *idaData) (newMaxScore int, found bool) {
	currentState := data.states[len(data.states)-1]
	score := data.fx(currentState, data.states[0], data.goal, data.path)
	data.tries++
	if currentComplexity := len(data.states); currentComplexity > data.closedSetComplexity {
		data.closedSetComplexity = currentComplexity
	}
	if score > data.maxScore {
		return score, false
	}
	if isEqual(currentState, data.goal) {
		return -1, true
	}
	minScoreAboveCutOff := 1 << 30
	for _, dir := range directions {
		ok, nextPos := dir.fx(currentState)
		if !ok {
			continue
		}
		nextHash, _, _ := matrixToStringSelector(nextPos, 1, 1)
		if index := Index(data.hashes, nextHash); index != -1 {
			continue
		}
		data.path = append(data.path, dir.name)
		data.states = append(data.states, nextPos)
		data.hashes = append(data.hashes, nextHash)

		newMaxScore, found := IDA(data)
		if found {
			return newMaxScore, true
		}
		if newMaxScore < minScoreAboveCutOff {
			minScoreAboveCutOff = newMaxScore
		}
		data.path = data.path[:len(data.path)-1]
		data.states = data.states[:len(data.states)-1]
		data.hashes = data.hashes[:len(data.hashes)-1]
	}
	return minScoreAboveCutOff, false
}
