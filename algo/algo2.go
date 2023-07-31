package algo

import (
	"fmt"
	"os"
)

func initDataIDA(param AlgoParameters) (data idaData) {
	data.MaxScore = param.Eval.Fx(param.Board, param.Board, Goal(len(param.Board)), []byte{})
	data.States = append(data.States, Deep2DSliceCopy(param.Board))
	hash, _, _ := MatrixToStringSelector(param.Board, 1, 1)
	data.Hashes = append(data.Hashes, hash)
	data.Fx = param.Eval.Fx
	data.Goal = Goal(len(param.Board))
	return
}

func iterateIDA(data *idaData) (result Result) {
	fmt.Fprintln(os.Stderr, "Selected ALGO : IDA*")
	for data.MaxScore < 1<<30 {
		fmt.Fprintln(os.Stderr, "Cut off is now :", data.MaxScore)
		newMaxScore, found := ida(data)
		if found {
			return Result{data.Path, data.ClosedSetComplexity, data.Tries, data.RamFailure, "IDA"}
		}
		data.MaxScore = newMaxScore
	}
	data.Path = nil
	return
}

func ida(data *idaData) (newMaxScore int, found bool) {
	currentState := data.States[len(data.States)-1]
	score := data.Fx(currentState, data.States[0], data.Goal, data.Path)
	data.Tries++
	if data.Tries > 0 && data.Tries%100000 == 0 {
		fmt.Fprintf(os.Stderr, "%d * 100k tries\n", data.Tries/100000)
	}
	if currentComplexity := len(data.States); currentComplexity > data.ClosedSetComplexity {
		data.ClosedSetComplexity = currentComplexity
	}
	if score > data.MaxScore {
		return score, false
	}
	if isEqual(currentState, data.Goal) {
		return -1, true
	}
	minScoreAboveCutOff := 1 << 30
	for _, dir := range Directions {
		if len(data.Path) > 0 {
			conflictStr := string(data.Path[len(data.Path)-1]) + string(dir.name)
			if conflictStr == "LR" || conflictStr == "RL" || conflictStr == "UD" || conflictStr == "DU" {
				continue
			}
		}
		ok, nextPos := dir.fx(currentState)
		if !ok {
			continue
		}
		nextHash, _, _ := MatrixToStringSelector(nextPos, 1, 1)
		if index := Index(data.Hashes, nextHash); index != -1 {
			continue
		}
		data.Path = append(data.Path, dir.name)
		data.States = append(data.States, nextPos)
		data.Hashes = append(data.Hashes, nextHash)

		newMaxScore, found := ida(data)
		if found {
			return newMaxScore, true
		}
		if newMaxScore < minScoreAboveCutOff {
			minScoreAboveCutOff = newMaxScore
		}
		data.Path = data.Path[:len(data.Path)-1]
		data.States = data.States[:len(data.States)-1]
		data.Hashes = data.Hashes[:len(data.Hashes)-1]
	}
	return minScoreAboveCutOff, false
}
