package algo

import (
	"os"
	"sync"
)

type Move2D struct {
	dir byte
	X   int
	Y   int
}

type Pos2D struct {
	X int
	Y int
}

type Node struct {
	//world [][]int
	world uint64
	path  []byte
	score int
}

func BoardToUint64(board [][]int) (res uint64) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			res |= uint64(board[i][j])
			if i != 3 || j != 3 {
				res <<= 4
			}
		}
	}
	return
}

func Uint64ToBoard(flat uint64) (board [][]int) {
	board = make([][]int, 4)
	for i := 3; i >= 0; i-- {
		board[i] = make([]int, 4)
		for j := 3; j >= 0; j-- {
			board[i][j] = int(flat & 15)
			flat >>= 4
		}
	}
	return
}

type moveFx func([][]int) (bool, [][]int)

type EvalFx func(pos, startPos, goalPos [][]int, path []byte) int

type Eval struct {
	Name string
	Fx   EvalFx
}

type Option struct {
	Filename         string
	Fd               *os.File
	MapSize          int
	Heuristic        string
	Workers          int
	SeenNodesSplit   int
	SpeedDisplay     int
	NoIterativeDepth bool
	Debug            bool
	DisableUI        bool
	StringInput      string
	RAMMaxGB         uint64
}

type Result struct {
	Path                []byte
	ClosedSetComplexity int
	Tries               int
	RamFailure          bool
	Algo                string
}

type idaData struct {
	Fx                  EvalFx
	MaxScore            int
	Path                []byte
	States              [][][]int
	Hashes              []string
	Goal                [][]int
	ClosedSetComplexity int
	Tries               int
	RamFailure          bool
}

type safeData struct {
	MuQueue  []sync.Mutex
	PosQueue []*PriorityQueue

	MuSeen       []sync.Mutex
	SeenNodes    []map[string]int
	Tries        int
	MaxSizeQueue []int

	Mu                  sync.Mutex
	Path                []byte
	Over                bool
	Win                 bool
	RamFailure          bool
	WinScore            int
	Idle                int
	ClosedSetComplexity int
	RAMMin              uint64
}

type AlgoParameters struct {
	Workers        int
	SeenNodesSplit int
	Eval           Eval
	Board          [][]int
	Unsolvable     bool
	RAMMaxGB       uint64
}
