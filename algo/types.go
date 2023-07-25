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
	world [][]int
	path  []byte
	score int
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
}

type Result struct {
	Path                []byte
	ClosedSetComplexity int
	Tries               int
	RamFailure          bool
	Algo				string
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
}

type AlgoParameters struct {
	Workers        int
	SeenNodesSplit int
	Eval           Eval
	Board          [][]int
	Unsolvable     bool
}
