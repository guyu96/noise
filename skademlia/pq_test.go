package skademlia

import (
	"container/heap"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPQInit(t *testing.T) {
	n := 100
	pq := make(PriorityQueue, n)
	for i := 0; i < n; i++ {
		pq[i] = &Item{
			value:    struct{}{},
			priority: rand.Uint64(),
			index:    i,
		}
	}
	heap.Init(&pq)

	//pq.update(item, item.value, 5)

	// decreasing priority order.
	prevPriority := ^uint64(0) // max uint
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		assert.True(t, prevPriority > item.priority)
	}
}

func TestPQPush(t *testing.T) {
	n := 20
	pq := make(PriorityQueue, n)
	for i := 0; i < n; i++ {
		pq[i] = &Item{
			value:    struct{}{},
			priority: rand.Uint64(),
			index:    i,
		}
	}
	heap.Init(&pq)

	for i := 0; i < n; i++ {
		item := &Item{
			value:    struct{}{},
			priority: rand.Uint64(),
		}
		pq.Push(item)
	}

	//pq.update(item, item.value, 5)

	// decreasing priority order.
	prevPriority := ^uint64(0) // max uint
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		assert.True(t, prevPriority > item.priority)
	}
}
