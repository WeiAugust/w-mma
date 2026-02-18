package ingest

import "context"

// Parser parses a remote URL into a pending review record.
type Parser interface {
	Parse(ctx context.Context, job FetchJob) (PendingRecord, error)
}

type Worker struct {
	queue  Queue
	repo   Repository
	parser Parser
}

func NewWorker(queue Queue, repo Repository, parser Parser) *Worker {
	return &Worker{queue: queue, repo: repo, parser: parser}
}

func NewQueuelessWorker(repo Repository, parser Parser) *Worker {
	return &Worker{repo: repo, parser: parser}
}

func (w *Worker) RunOnce(ctx context.Context) {
	if w.queue == nil {
		return
	}
	job, ok := w.queue.Pop(ctx)
	if !ok {
		return
	}
	_ = w.HandleJob(ctx, job)
}

func (w *Worker) HandleJob(ctx context.Context, job FetchJob) error {
	rec, err := w.parser.Parse(ctx, job)
	if err != nil {
		return err
	}
	rec.SourceID = job.SourceID
	return w.repo.SavePending(ctx, rec)
}
