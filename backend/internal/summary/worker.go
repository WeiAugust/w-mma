package summary

import "context"

type Worker struct {
	repo      Repository
	summarize Summarizer
}

func NewWorker(repo Repository, summarize Summarizer) *Worker {
	return &Worker{repo: repo, summarize: summarize}
}

func (w *Worker) RunOnce(ctx context.Context, jobID int64, content string) error {
	if err := w.repo.UpdateStatus(ctx, jobID, StatusRunning, ""); err != nil {
		return err
	}

	if _, err := w.summarize.Summarize(ctx, content); err != nil {
		_ = w.repo.UpdateStatus(ctx, jobID, StatusFailed, err.Error())
		return err
	}

	return w.repo.UpdateStatus(ctx, jobID, StatusDone, "")
}
