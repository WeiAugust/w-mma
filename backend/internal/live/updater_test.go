package live

import (
	"context"
	"testing"
)

type fakeLiveRepo struct {
	updateCount int
}

func newFakeLiveRepo() *fakeLiveRepo {
	return &fakeLiveRepo{}
}

func (r *fakeLiveRepo) UpsertBoutResult(context.Context, BoutResult) error {
	r.updateCount++
	return nil
}

type fakeLiveClient struct {
	winner string
}

func fakeLiveClientWinner(winner string) fakeLiveClient {
	return fakeLiveClient{winner: winner}
}

func (c fakeLiveClient) FetchEventResults(context.Context, int64) ([]BoutResult, error) {
	return []BoutResult{{BoutID: 1001, Winner: c.winner, Method: "KO", Round: 2}}, nil
}

func TestUpdater_UpdatesBoutResultIdempotently(t *testing.T) {
	repo := newFakeLiveRepo()
	client := fakeLiveClientWinner("fighter_a")
	u := NewUpdater(repo, client)

	_ = u.UpdateEvent(context.Background(), 10)
	_ = u.UpdateEvent(context.Background(), 10)

	if repo.updateCount != 1 {
		t.Fatalf("expected idempotent update count 1, got %d", repo.updateCount)
	}
}
