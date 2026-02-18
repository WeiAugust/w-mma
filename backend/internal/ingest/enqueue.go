package ingest

import "context"

// FetchJob describes one source URL fetch request.
type FetchJob struct {
	SourceID   int64
	URL        string
	ParserKind string
}

// Queue is the ingest job queue abstraction.
type Queue interface {
	Push(job FetchJob)
	Pop(ctx context.Context) (FetchJob, bool)
}

// Enqueuer sends fetch jobs to the queue.
type Enqueuer struct {
	queue Queue
}

func NewEnqueuer(queue Queue) *Enqueuer {
	return &Enqueuer{queue: queue}
}

func (e *Enqueuer) Enqueue(job FetchJob) {
	e.queue.Push(job)
}
