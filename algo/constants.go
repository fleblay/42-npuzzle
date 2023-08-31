package algo

var Evals = []Eval{
	{"dijkstra", dijkstra},
	{"greedy_hamming", greedy_hamming},
	{"greedy_manhattan", greedy_manhattan},
	{"astar_hamming", astar_hamming},
	{"astar_manhattan", astar_manhattan_generator(1)},
	{"astar_manhattan2", astar_manhattan_generator(2)},
	{"astar_manhattan1.3", astar_manhattan_generator(1.3)},
	{"astar_manhattan_conflict", astar_manhattan_generator_conflict(1)},
	{"astar_manhattan_conflict1.3", astar_manhattan_generator_conflict(1.3)},
}

var Directions = []struct {
	name byte
	fx   moveFx
}{
	{'U', moveUp},
	{'D', moveDown},
	{'L', moveLeft},
	{'R', moveRight},
}

var MinRAMAvailableMB uint64 = 256
