package main

import (
	"container/heap"
	"fmt"
	"os"
)

type Item struct {
	node  Node
	index int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].node.score < pq[j].node.score
}

func (pq PriorityQueue) Swap(i, j int) {

	if i < 0 || j < 0 {
		fmt.Fprintf(os.Stderr, "swap with index i:%d, j:%d. Len is %d\n", i, j, pq.Len())
	}
	pq[i], pq[j] = pq[j], pq[i]
	if i < 0 || j < 0 {
		fmt.Fprintf(os.Stderr, "swap OK with index i:%d, j:%d. Len is %d\n", i, j, pq.Len())
	}
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	newItem := x.(*Item)
	newItem.index = n
	*pq = append(*pq, newItem)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) update(item *Item, updatedNode Node) {
	item.node = updatedNode
	heap.Fix(pq, item.index)
}
