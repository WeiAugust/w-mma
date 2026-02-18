package ingest

import (
	"context"
	"errors"
	"testing"
)

type fakeQueue struct {
	items []FetchJob
}

func newFakeQueue() *fakeQueue {
	return &fakeQueue{}
}

func (q *fakeQueue) Push(job FetchJob) {
	q.items = append(q.items, job)
}

func (q *fakeQueue) Pop(context.Context) (FetchJob, bool) {
	if len(q.items) == 0 {
		return FetchJob{}, false
	}
	job := q.items[0]
	q.items = q.items[1:]
	return job, true
}

type fakeRepo struct {
	pendingCount int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{}
}

func (r *fakeRepo) SavePending(context.Context, PendingRecord) error {
	r.pendingCount++
	return nil
}

type fakeParser struct {
	err error
}

func fakeParserSuccess() fakeParser {
	return fakeParser{}
}

func (p fakeParser) Parse(context.Context, FetchJob) (PendingRecord, error) {
	if p.err != nil {
		return PendingRecord{}, p.err
	}
	return PendingRecord{Title: "news-a", SourceURL: "https://example.com/a"}, nil
}

func TestWorker_StoresPendingReviewOnSuccess(t *testing.T) {
	queue := newFakeQueue()
	repo := newFakeRepo()
	queue.Push(FetchJob{SourceID: 1, URL: "https://example.com/a"})

	w := NewWorker(queue, repo, fakeParserSuccess())
	w.RunOnce(context.Background())

	if repo.pendingCount != 1 {
		t.Fatalf("expected 1 pending record, got %d", repo.pendingCount)
	}
}

func TestWorker_IgnoresParserFailure(t *testing.T) {
	queue := newFakeQueue()
	repo := newFakeRepo()
	queue.Push(FetchJob{SourceID: 1, URL: "https://example.com/a"})

	w := NewWorker(queue, repo, fakeParser{err: errors.New("boom")})
	w.RunOnce(context.Background())

	if repo.pendingCount != 0 {
		t.Fatalf("expected 0 pending record, got %d", repo.pendingCount)
	}
}
