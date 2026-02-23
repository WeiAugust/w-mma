package live

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/ufc"
)

type fakeUFCLiveRepo struct {
	events             []UFCTrackableEvent
	boutsByEvent       map[int64][]UFCBoutSnapshot
	eventStatusUpdates []string
	boutUpdates        int
}

func (r *fakeUFCLiveRepo) ListTrackableEvents(context.Context) ([]UFCTrackableEvent, error) {
	items := make([]UFCTrackableEvent, len(r.events))
	copy(items, r.events)
	return items, nil
}

func (r *fakeUFCLiveRepo) ListBoutSnapshots(_ context.Context, eventID int64) ([]UFCBoutSnapshot, error) {
	items := make([]UFCBoutSnapshot, len(r.boutsByEvent[eventID]))
	copy(items, r.boutsByEvent[eventID])
	return items, nil
}

func (r *fakeUFCLiveRepo) UpdateEventStatus(_ context.Context, eventID int64, status string) error {
	r.eventStatusUpdates = append(r.eventStatusUpdates, status)
	for i := range r.events {
		if r.events[i].ID == eventID {
			r.events[i].Status = status
		}
	}
	return nil
}

func (r *fakeUFCLiveRepo) UpsertBoutResult(_ context.Context, eventID int64, boutID int64, winnerID int64, method string, round int, timeSec int, result string) error {
	r.boutUpdates++
	for i := range r.boutsByEvent[eventID] {
		item := &r.boutsByEvent[eventID][i]
		if item.BoutID != boutID {
			continue
		}
		item.WinnerID = winnerID
		item.Method = method
		item.Round = round
		item.TimeSec = timeSec
		item.Result = result
		return nil
	}
	return nil
}

type fakeUFCCache struct {
	invalidatedEventIDs []int64
	invalidatedEvents   int
}

func (c *fakeUFCCache) InvalidateEvent(_ context.Context, eventID int64) error {
	c.invalidatedEventIDs = append(c.invalidatedEventIDs, eventID)
	return nil
}

func (c *fakeUFCCache) InvalidateEvents(context.Context) error {
	c.invalidatedEvents++
	return nil
}

type fakeUFCScraper struct {
	card  ufc.EventCard
	err   error
	calls int
}

func (s *fakeUFCScraper) GetEventCard(context.Context, string) (ufc.EventCard, error) {
	s.calls++
	if s.err != nil {
		return ufc.EventCard{}, s.err
	}
	return s.card, nil
}

func TestUFCLiveMonitor_TransitionsScheduledEventToLive(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	repo := &fakeUFCLiveRepo{
		events: []UFCTrackableEvent{
			{ID: 10, Status: "scheduled", StartsAt: now.Add(-time.Minute), ExternalURL: "https://www.ufc.com/event/ufc-326"},
		},
		boutsByEvent: map[int64][]UFCBoutSnapshot{},
	}
	cache := &fakeUFCCache{}
	scraper := &fakeUFCScraper{}
	monitor := NewUFCLiveMonitor(repo, scraper, cache, UFCLiveMonitorConfig{
		TickInterval:     time.Minute,
		MinPollInterval:  5 * time.Minute,
		MaxPollInterval:  5 * time.Minute,
		MaxPollPerTick:   1,
		Random:           rand.New(rand.NewSource(7)),
		Now:              func() time.Time { return now },
		RetryBackoffPlan: []time.Duration{10 * time.Minute, 20 * time.Minute, 40 * time.Minute},
	})

	if err := monitor.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(repo.eventStatusUpdates) == 0 || repo.eventStatusUpdates[0] != "live" {
		t.Fatalf("expected scheduled event to transition to live, got updates=%v", repo.eventStatusUpdates)
	}
}

func TestUFCLiveMonitor_UpdatesBoutAndCompletesEvent(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	repo := &fakeUFCLiveRepo{
		events: []UFCTrackableEvent{
			{ID: 10, Status: "live", StartsAt: now.Add(-20 * time.Minute), ExternalURL: "https://www.ufc.com/event/ufc-326"},
		},
		boutsByEvent: map[int64][]UFCBoutSnapshot{
			10: {
				{BoutID: 1001, SequenceNo: 1, RedFighterID: 20, BlueFighterID: 21},
			},
		},
	}
	cache := &fakeUFCCache{}
	scraper := &fakeUFCScraper{
		card: ufc.EventCard{
			Status: "completed",
			Bouts: []ufc.EventBout{
				{WinnerSide: "red", Method: "KO/TKO", Round: 2, TimeSec: 100, Result: "KO/TKO R2 1:40"},
			},
		},
	}
	monitor := NewUFCLiveMonitor(repo, scraper, cache, UFCLiveMonitorConfig{
		TickInterval:     time.Minute,
		MinPollInterval:  time.Second,
		MaxPollInterval:  time.Second,
		MaxPollPerTick:   1,
		Random:           rand.New(rand.NewSource(7)),
		Now:              func() time.Time { return now },
		RetryBackoffPlan: []time.Duration{10 * time.Minute, 20 * time.Minute, 40 * time.Minute},
	})
	monitor.nextCheckAt[10] = now

	if err := monitor.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if repo.boutUpdates != 1 {
		t.Fatalf("expected 1 bout update, got %d", repo.boutUpdates)
	}
	if repo.boutsByEvent[10][0].WinnerID != 20 {
		t.Fatalf("expected winner id 20, got %d", repo.boutsByEvent[10][0].WinnerID)
	}
	if got := repo.boutsByEvent[10][0].Method; got != "KO/TKO" {
		t.Fatalf("expected method KO/TKO, got %q", got)
	}
	if _, ok := monitor.nextCheckAt[10]; ok {
		t.Fatalf("expected completed event to stop tracking")
	}
}

func TestUFCLiveMonitor_FailureUsesBackoff(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	repo := &fakeUFCLiveRepo{
		events: []UFCTrackableEvent{
			{ID: 10, Status: "live", StartsAt: now.Add(-20 * time.Minute), ExternalURL: "https://www.ufc.com/event/ufc-326"},
		},
		boutsByEvent: map[int64][]UFCBoutSnapshot{
			10: {{BoutID: 1001, SequenceNo: 1, RedFighterID: 20, BlueFighterID: 21}},
		},
	}
	cache := &fakeUFCCache{}
	scraper := &fakeUFCScraper{err: errors.New("429 too many requests")}
	monitor := NewUFCLiveMonitor(repo, scraper, cache, UFCLiveMonitorConfig{
		TickInterval:     time.Minute,
		MinPollInterval:  time.Second,
		MaxPollInterval:  time.Second,
		MaxPollPerTick:   1,
		Random:           rand.New(rand.NewSource(7)),
		Now:              func() time.Time { return now },
		RetryBackoffPlan: []time.Duration{10 * time.Minute, 20 * time.Minute, 40 * time.Minute},
	})
	monitor.nextCheckAt[10] = now

	if err := monitor.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if got := monitor.nextCheckAt[10]; !got.Equal(now.Add(10 * time.Minute)) {
		t.Fatalf("expected first backoff to 10m, got %s", got.Sub(now))
	}
}

func TestUFCLiveMonitor_StaleLiveEventForceCompletesWithoutPolling(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	repo := &fakeUFCLiveRepo{
		events: []UFCTrackableEvent{
			{ID: 10, Status: "live", StartsAt: now.Add(-48 * time.Hour), ExternalURL: "https://www.ufc.com/event/ufc-323"},
		},
		boutsByEvent: map[int64][]UFCBoutSnapshot{
			10: {{BoutID: 1001, SequenceNo: 1, RedFighterID: 20, BlueFighterID: 21}},
		},
	}
	cache := &fakeUFCCache{}
	scraper := &fakeUFCScraper{err: errors.New("should not be called")}
	monitor := NewUFCLiveMonitor(repo, scraper, cache, UFCLiveMonitorConfig{
		TickInterval:     time.Minute,
		MinPollInterval:  time.Second,
		MaxPollInterval:  time.Second,
		MaxPollPerTick:   1,
		Random:           rand.New(rand.NewSource(7)),
		Now:              func() time.Time { return now },
		RetryBackoffPlan: []time.Duration{10 * time.Minute, 20 * time.Minute, 40 * time.Minute},
	})

	if err := monitor.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(repo.eventStatusUpdates) == 0 || repo.eventStatusUpdates[0] != "completed" {
		t.Fatalf("expected stale live event to be completed, got updates=%v", repo.eventStatusUpdates)
	}
	if scraper.calls != 0 {
		t.Fatalf("expected stale live event to skip polling, got calls=%d", scraper.calls)
	}
}
