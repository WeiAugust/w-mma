package live

import (
	"context"
	"fmt"
	"sync"
)

// BoutResult represents one bout result snapshot.
type BoutResult struct {
	BoutID  int64  `json:"bout_id"`
	Winner  string `json:"winner"`
	Method  string `json:"method"`
	Round   int    `json:"round"`
	TimeSec int    `json:"time_sec"`
}

type Repository interface {
	UpsertBoutResult(ctx context.Context, result BoutResult) error
}

type Client interface {
	FetchEventResults(ctx context.Context, eventID int64) ([]BoutResult, error)
}

type Updater struct {
	repo   Repository
	client Client

	mu   sync.Mutex
	seen map[string]struct{}
}

func NewUpdater(repo Repository, client Client) *Updater {
	return &Updater{repo: repo, client: client, seen: map[string]struct{}{}}
}

func (u *Updater) UpdateEvent(ctx context.Context, eventID int64) error {
	results, err := u.client.FetchEventResults(ctx, eventID)
	if err != nil {
		return err
	}

	for _, r := range results {
		key := fmt.Sprintf("%d:%d:%s:%s:%d:%d", eventID, r.BoutID, r.Winner, r.Method, r.Round, r.TimeSec)
		if !u.markSeen(key) {
			continue
		}
		if err := u.repo.UpsertBoutResult(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) markSeen(key string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.seen[key]; ok {
		return false
	}
	u.seen[key] = struct{}{}
	return true
}
