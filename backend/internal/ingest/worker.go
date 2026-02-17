package ingest

import "context"

// Parser parses a remote URL into a pending review record.
type Parser interface {
	Parse(ctx context.Context, url string) (PendingRecord, error)
}

type Worker struct {
	queue  Queue
	repo   Repository
	parser Parser
}

func NewWorker(queue Queue, repo Repository, parser Parser) *Worker {
	return &Worker{queue: queue, repo: repo, parser: parser}
}

func (w *Worker) RunOnce(ctx context.Context) {
	job, ok := w.queue.Pop(ctx)
	if !ok {
		return
	}

	rec, err := w.parser.Parse(ctx, job.URL)
	if err != nil {
		return
	}
	rec.SourceID = job.SourceID
	_ = w.repo.SavePending(ctx, rec)
}
