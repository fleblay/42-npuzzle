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
	world uint64 // max size is now 4
	path  []byte // could be replaced with [5]uint64 (80moves of 2bit) + uint8 for len of path
	score uint16
}

//TODO fx to change for 3x3 or 4x4

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

//TODO fx to change for 3x3 or 4x4
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
	//Hashes              []string
	Hashes              []uint64
	Goal                [][]int
	ClosedSetComplexity int
	Tries               int
	RamFailure          bool
}

type safeData struct {
	MuQueue  []sync.Mutex
	PosQueue []*PriorityQueue

	MuSeen       []sync.Mutex
	//SeenNodes    []map[string]int
	SeenNodes    []map[uint64]int
	Tries        int
	MaxSizeQueue []int

	Mu                  sync.Mutex
	Path                []byte
	Over                bool
	Win                 bool
	RamFailure          bool
	WinScore            uint16
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
