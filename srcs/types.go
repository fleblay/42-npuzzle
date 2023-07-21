package main

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

type evalFx func(pos, startPos, goalPos [][]int, path []byte) int

type eval struct {
	name string
	fx   evalFx
}
type option struct {
	filename         string
	fd               *os.File
	mapSize          int
	heuristic        string
	workers          int
	seenNodesSplit   int
	speedDisplay     int
	noIterativeDepth bool
	debug            bool
	disableUI        bool
}

type algoParameters struct {
	workers        int
	seenNodesSplit int
	maxScore       int
	eval           eval
	board          [][]int
	unsolvable     bool
}

type safeData struct {
	muQueue  []sync.Mutex
	posQueue []*PriorityQueue

	muSeen       []sync.Mutex
	seenNodes    []map[string]int
	tries        int
	maxSizeQueue []int

	mu                  sync.Mutex
	path                []byte
	over                bool
	win                 bool
	ramFailure			bool
	winScore            int
	idle                int
	closedSetComplexity int
}
