package cdcl_solver

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/styczynski/go-sat-solver/sat_solver"
)

/**
 * Single item inside the literal priority queue
 * The queue will compare items by solver score saved in activity field.
 * Higher scores are on the top.
 */
type PQLitItem struct {
	// Literal
	value sat_solver.CNFLiteral
	// Position of this literal inside the heap
	index int
}

/**
 * Copy the item
 */
func (pqItem *PQLitItem) Copy() *PQLitItem {
	return &PQLitItem{
		value: pqItem.value,
		index: pqItem.index,
	}
}

/*
 * LiteralPriorityQueue implements heap.Interface and stores literals comparing them by solver.activity.
 * For more details how priority queues can be implemented in Go please see:
 *   https://golang.org/src/container/heap/example_pq_test.go
 */
type LiteralPriorityQueue struct {
	// Solver instance
	solver *CDCLSolver
	// Items on a heap
	items []*PQLitItem
	// Mapping from literal to its index
	indexes map[sat_solver.CNFLiteral]int
}

/**
 * Make a copy of the queue. Useful when you want to examine few max items without messing up the queue.
 */
func (pq LiteralPriorityQueue) Copy() LiteralPriorityQueue {
	ret := LiteralPriorityQueue{
		solver:  pq.solver,
		items:   make([]*PQLitItem, len(pq.items)),
		indexes: map[sat_solver.CNFLiteral]int{},
	}
	for i, item := range pq.items {
		ret.items[i] = item.Copy()
		ret.indexes[ret.items[i].value] = i
	}
	return ret
}

/**
 * Human readable dump of the queue.
 */
func (pq LiteralPriorityQueue) String() string {
	c := pq.Copy()
	ret := make([]string, c.Len())
	i := 0
	for c.Len() > 0 {
		item := heap.Pop(&c).(*PQLitItem)
		ret[i] = fmt.Sprintf("%s(%.2f)", item.value.DebugString(), pq.solver.activity[item.value])
		i++
	}
	return fmt.Sprintf("Queue[%s]", strings.Join(ret, ", "))
}

/**
 * Get length of the queue.
 */
func (pq LiteralPriorityQueue) Len() int {
	return len(pq.items)
}

/**
 * Check if the queue contains literal. Executes in O(1).
 */
func (pq LiteralPriorityQueue) Has(lit sat_solver.CNFLiteral) bool {
	 _, ok := pq.indexes[lit]
	 return ok
}

/**
 * Literals comparator.
 */
func (pq LiteralPriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq.solver.activity[pq.items[i].value] > pq.solver.activity[pq.items[j].value]
}

/**
 * Used to swap literals on two indices.
 */
func (pq LiteralPriorityQueue) Swap(i, j int) {
	// Move items
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	// Update stored indexes
	pq.items[i].index = i
	pq.items[j].index = j
	// Update index mapping
	pq.indexes[pq.items[i].value] = i
	pq.indexes[pq.items[j].value] = j
}

/**
 * Insert new item on the heap.
 * Please note that the argument is *PQLitItem and not CNFLiteral!
 */
func (pq *LiteralPriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*PQLitItem)
	item.index = n
	pq.indexes[item.value] = n
	pq.items = append(pq.items, item)
}

/**
 * Remove the top item from the queue and return it.
 * Returns *PQLitItem and not the CNFLiteral, so be careful!
 */
func (pq *LiteralPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old.items)
	item := old.items[n-1]
	old.items[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	delete(pq.indexes, item.value)
	pq.items = old.items[0 : n-1]
	return item
}

/**
 * Forces the queue to fix iteself after you change priority for the given literal.
 */
func (pq *LiteralPriorityQueue) Update(value sat_solver.CNFLiteral) {
	if i, ok := pq.indexes[value]; ok {
		heap.Fix(pq, i)
	}
}

/**
 * Create new priority queue for a given solver instance.
 */
func NewLiteralPriorityQueue(solver *CDCLSolver) *LiteralPriorityQueue {
	vars := solver.vars.GetAllVariables()
	pq := LiteralPriorityQueue{
		solver: solver,
		items:  make([]*PQLitItem, len(vars)),
		indexes: map[sat_solver.CNFLiteral]int{},
	}
	for i, v := range vars {
		if v < 0 {
			v = -v
		}
		pq.items[i] = &PQLitItem{
			value: v,
			index: i,
		}
		pq.indexes[v] = i
	}
	heap.Init(&pq)
	return &pq
}