package ingest

import (
	"context"
	"sync"
)

// MemoryQueue is an in-memory queue for local development/testing.
type MemoryQueue struct {
	mu    sync.Mutex
	items []FetchJob
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{items: []FetchJob{}}
}

func (q *MemoryQueue) Push(job FetchJob) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, job)
}

func (q *MemoryQueue) Pop(context.Context) (FetchJob, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return FetchJob{}, false
	}
	job := q.items[0]
	q.items = q.items[1:]
	return job, true
}
