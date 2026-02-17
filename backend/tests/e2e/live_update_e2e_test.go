package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/event"
	"github.com/bajiaozhi/w-mma/backend/internal/live"
)

type fakeLiveClient struct {
	mu     sync.Mutex
	winner string
}

func (c *fakeLiveClient) SetWinner(winner string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.winner = winner
}

func (c *fakeLiveClient) FetchEventResults(context.Context, int64) ([]live.BoutResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.winner == "" {
		return []live.BoutResult{}, nil
	}
	return []live.BoutResult{{
		BoutID: 1001,
		Winner: c.winner,
		Method: "KO",
		Round:  2,
	}}, nil
}

type liveEventRepoAdapter struct {
	eventRepo *event.InMemoryRepository
}

func (a *liveEventRepoAdapter) UpsertBoutResult(ctx context.Context, result live.BoutResult) error {
	winnerID, _ := strconv.ParseInt(result.Winner, 10, 64)
	return a.eventRepo.UpsertBoutResult(ctx, 10, result.BoutID, winnerID, result.Method)
}

func TestE2E_LiveEventUpdatesEvery30Seconds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eventRepo := event.NewInMemoryRepository()
	eventSvc := event.NewService(eventRepo)

	r := gin.New()
	event.RegisterEventRoutes(r, eventSvc)
	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &fakeLiveClient{}
	updater := live.NewUpdater(&liveEventRepoAdapter{eventRepo: eventRepo}, client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go live.RunScheduler(ctx, updater, 10)

	time.AfterFunc(1*time.Second, func() {
		client.SetWinner("20")
	})

	deadline := time.Now().Add(35 * time.Second)
	for time.Now().Before(deadline) {
		res, err := http.Get(ts.URL + "/api/events/10")
		if err != nil {
			t.Fatalf("request event card failed: %v", err)
		}

		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()

		var card struct {
			Bouts []struct {
				WinnerID int64 `json:"winner_id"`
			} `json:"bouts"`
		}
		if err := json.Unmarshal(body, &card); err != nil {
			t.Fatalf("decode event response failed: %v", err)
		}

		if len(card.Bouts) > 0 && card.Bouts[0].WinnerID == 20 {
			return
		}
		time.Sleep(1 * time.Second)
	}

	t.Fatalf("expected winner update within 35 seconds")
}
