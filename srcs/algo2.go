package main

import (
	"fmt"
	"os"
)

func initData2(param algoParameters) (data idaData) {
	data.maxScore = param.eval.fx(param.board, param.board, goal(len(param.board)), []byte{}) + 1
	data.states = append(data.states, Deep2DSliceCopy(param.board))
	hash, _, _ := matrixToStringSelector(param.board, 1, 1)
	data.hashes = append(data.hashes, hash)
	data.fx = param.eval.fx
	data.goal = goal(len(param.board))
	return
}

func iterateAlgo2(data *idaData) {
	for data.maxScore < 1<<30 {
		fmt.Fprintln(os.Stderr, "true IDA* cut off is now :", data.maxScore)
		newMaxScore, found := algo2(data)
		if found {
			fmt.Fprintln(os.Stderr, "Solution !")
			return
		}
		data.maxScore = newMaxScore
	}
	fmt.Fprintln(os.Stderr, "No solution")
	data.path = nil
}

func algo2(data *idaData) (newMaxScore int, found bool) {
	//fmt.Fprintln(os.Stderr, "Deeper")
	currentState := data.states[len(data.states)-1]
	score := data.fx(currentState, data.states[0], data.goal, data.path)
	if score > data.maxScore {
		//fmt.Fprintln(os.Stderr, "Score")
		return score, false
	}
	if isEqual(currentState, data.goal) {
		//fmt.Fprintln(os.Stderr, "Solution !!!")
		return -1, true
	}
	minScoreAboveCutOff := 1 << 30 
	for _, dir := range directions {
		ok, nextPos := dir.fx(currentState)
		if !ok {
			//fmt.Fprintln(os.Stderr, string(dir.name), "is not ok")
			continue
		}
		nextHash, _, _ := matrixToStringSelector(nextPos, 1, 1)
		if index := Index(data.hashes, nextHash); index != -1 {
			//fmt.Fprintln(os.Stderr, "hash already seen")
			continue
		}
		data.path = append(data.path, dir.name)
		data.states = append(data.states, nextPos)
		data.hashes = append(data.hashes, nextHash)

		newMaxScore, found := algo2(data)
		//fmt.Fprintln(os.Stderr, "after return of algo", newMaxScore, found)
		if found {
			return newMaxScore, true
		}
		if newMaxScore < minScoreAboveCutOff {
			//fmt.Fprintln(os.Stderr, "min is now", minScoreAboveCutOff)
			minScoreAboveCutOff = newMaxScore
		}
		data.path = data.path[:len(data.path)-1]
		data.states = data.states[:len(data.states)-1]
		data.hashes = data.hashes[:len(data.hashes)-1]
	}
	//fmt.Fprintln(os.Stderr, "About to return with score ", minScoreAboveCutOff)
	return minScoreAboveCutOff, false
}
