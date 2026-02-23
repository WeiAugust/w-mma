package live

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/ufc"
)

const staleLiveCompletionWindow = 18 * time.Hour

type UFCTrackableEvent struct {
	ID          int64
	Status      string
	StartsAt    time.Time
	ExternalURL string
}

type UFCBoutSnapshot struct {
	BoutID        int64
	SequenceNo    int
	RedFighterID  int64
	BlueFighterID int64
	WinnerID      int64
	Method        string
	Round         int
	TimeSec       int
	Result        string
}

type UFCEventRepository interface {
	ListTrackableEvents(ctx context.Context) ([]UFCTrackableEvent, error)
	ListBoutSnapshots(ctx context.Context, eventID int64) ([]UFCBoutSnapshot, error)
	UpdateEventStatus(ctx context.Context, eventID int64, status string) error
	UpsertBoutResult(ctx context.Context, eventID int64, boutID int64, winnerID int64, method string, round int, timeSec int, result string) error
}

type UFCEventScraper interface {
	GetEventCard(ctx context.Context, eventURL string) (ufc.EventCard, error)
}

type UFCEventCache interface {
	InvalidateEvent(ctx context.Context, eventID int64) error
	InvalidateEvents(ctx context.Context) error
}

type UFCLiveMonitorConfig struct {
	TickInterval     time.Duration
	MinPollInterval  time.Duration
	MaxPollInterval  time.Duration
	MaxPollPerTick   int
	RetryBackoffPlan []time.Duration
	Random           *rand.Rand
	Now              func() time.Time
}

type UFCLiveMonitor struct {
	repo    UFCEventRepository
	scraper UFCEventScraper
	cache   UFCEventCache

	tickInterval     time.Duration
	minPollInterval  time.Duration
	maxPollInterval  time.Duration
	maxPollPerTick   int
	retryBackoffPlan []time.Duration
	random           *rand.Rand
	now              func() time.Time

	nextCheckAt map[int64]time.Time
	failure     map[int64]int
}

func NewUFCLiveMonitor(repo UFCEventRepository, scraper UFCEventScraper, cache UFCEventCache, cfg UFCLiveMonitorConfig) *UFCLiveMonitor {
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = time.Minute
	}
	if cfg.MinPollInterval <= 0 {
		cfg.MinPollInterval = 5 * time.Minute
	}
	if cfg.MaxPollInterval <= 0 {
		cfg.MaxPollInterval = 10 * time.Minute
	}
	if cfg.MaxPollInterval < cfg.MinPollInterval {
		cfg.MaxPollInterval = cfg.MinPollInterval
	}
	if cfg.MaxPollPerTick <= 0 {
		cfg.MaxPollPerTick = 1
	}
	if len(cfg.RetryBackoffPlan) == 0 {
		cfg.RetryBackoffPlan = []time.Duration{10 * time.Minute, 20 * time.Minute, 40 * time.Minute}
	}
	if cfg.Random == nil {
		cfg.Random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	return &UFCLiveMonitor{
		repo:             repo,
		scraper:          scraper,
		cache:            cache,
		tickInterval:     cfg.TickInterval,
		minPollInterval:  cfg.MinPollInterval,
		maxPollInterval:  cfg.MaxPollInterval,
		maxPollPerTick:   cfg.MaxPollPerTick,
		retryBackoffPlan: cfg.RetryBackoffPlan,
		random:           cfg.Random,
		now:              cfg.Now,
		nextCheckAt:      map[int64]time.Time{},
		failure:          map[int64]int{},
	}
}

func (m *UFCLiveMonitor) Run(ctx context.Context) {
	_ = m.RunOnce(ctx)
	ticker := time.NewTicker(m.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.RunOnce(ctx)
		}
	}
}

func (m *UFCLiveMonitor) RunOnce(ctx context.Context) error {
	events, err := m.repo.ListTrackableEvents(ctx)
	if err != nil {
		return err
	}
	now := m.now().UTC()
	active := map[int64]struct{}{}
	for _, event := range events {
		active[event.ID] = struct{}{}
	}
	for eventID := range m.nextCheckAt {
		if _, ok := active[eventID]; !ok {
			delete(m.nextCheckAt, eventID)
			delete(m.failure, eventID)
		}
	}

	pollCount := 0
	for _, item := range events {
		if strings.TrimSpace(item.ExternalURL) == "" {
			continue
		}

		status := strings.ToLower(strings.TrimSpace(item.Status))
		if status == "scheduled" && !item.StartsAt.IsZero() && !item.StartsAt.After(now) {
			if err := m.repo.UpdateEventStatus(ctx, item.ID, "live"); err == nil {
				status = "live"
				m.invalidateCache(ctx, item.ID)
			}
		}
		if status != "live" {
			continue
		}
		if !item.StartsAt.IsZero() && item.StartsAt.Before(now.Add(-staleLiveCompletionWindow)) {
			if err := m.repo.UpdateEventStatus(ctx, item.ID, "completed"); err == nil {
				delete(m.nextCheckAt, item.ID)
				delete(m.failure, item.ID)
				m.invalidateCache(ctx, item.ID)
			}
			continue
		}

		dueAt, ok := m.nextCheckAt[item.ID]
		if !ok {
			m.nextCheckAt[item.ID] = now.Add(m.randomInterval())
			continue
		}
		if now.Before(dueAt) {
			continue
		}
		if pollCount >= m.maxPollPerTick {
			continue
		}
		pollCount++

		completed, err := m.pollLiveEvent(ctx, item, now)
		if err != nil {
			m.applyBackoff(item.ID, now)
			continue
		}
		if completed {
			delete(m.nextCheckAt, item.ID)
			delete(m.failure, item.ID)
			continue
		}

		m.failure[item.ID] = 0
		m.nextCheckAt[item.ID] = now.Add(m.randomInterval())
	}
	return nil
}

func (m *UFCLiveMonitor) pollLiveEvent(ctx context.Context, event UFCTrackableEvent, now time.Time) (bool, error) {
	card, err := m.scraper.GetEventCard(ctx, event.ExternalURL)
	if err != nil {
		return false, err
	}
	bouts, err := m.repo.ListBoutSnapshots(ctx, event.ID)
	if err != nil {
		return false, err
	}

	changed := false
	limit := len(card.Bouts)
	if len(bouts) < limit {
		limit = len(bouts)
	}
	for idx := 0; idx < limit; idx++ {
		src := card.Bouts[idx]
		current := bouts[idx]

		winnerID := winnerIDBySide(src.WinnerSide, current.RedFighterID, current.BlueFighterID)
		method := strings.TrimSpace(src.Method)
		round := src.Round
		timeSec := src.TimeSec
		result := strings.TrimSpace(src.Result)

		if winnerID == 0 && method == "" && round <= 0 && timeSec <= 0 && result == "" {
			continue
		}
		if winnerID == current.WinnerID &&
			method == strings.TrimSpace(current.Method) &&
			round == current.Round &&
			timeSec == current.TimeSec &&
			result == strings.TrimSpace(current.Result) {
			continue
		}
		if err := m.repo.UpsertBoutResult(ctx, event.ID, current.BoutID, winnerID, method, round, timeSec, result); err != nil {
			return false, err
		}
		changed = true
	}

	status := normalizePolledStatus(card.Status, event.StartsAt, now, card.Bouts)
	completed := status == "completed"
	if completed {
		if err := m.repo.UpdateEventStatus(ctx, event.ID, "completed"); err != nil {
			return false, err
		}
		changed = true
	}

	if changed {
		m.invalidateCache(ctx, event.ID)
	}
	return completed, nil
}

func winnerIDBySide(side string, redID int64, blueID int64) int64 {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "red":
		return redID
	case "blue":
		return blueID
	default:
		return 0
	}
}

func normalizePolledStatus(raw string, startsAt time.Time, now time.Time, bouts []ufc.EventBout) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "completed", "final":
		return "completed"
	case "live":
		if !startsAt.IsZero() && startsAt.Before(now.Add(-staleLiveCompletionWindow)) {
			return "completed"
		}
		return "live"
	case "scheduled", "upcoming":
		if startsAt.IsZero() || startsAt.After(now) {
			return "scheduled"
		}
		if startsAt.Before(now.Add(-staleLiveCompletionWindow)) {
			return "completed"
		}
		return "live"
	}
	if allBoutsResolved(bouts) {
		return "completed"
	}
	if startsAt.IsZero() {
		return "live"
	}
	if startsAt.After(now) {
		return "scheduled"
	}
	if startsAt.Before(now.Add(-staleLiveCompletionWindow)) {
		return "completed"
	}
	return "live"
}

func allBoutsResolved(items []ufc.EventBout) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if winnerIDBySide(item.WinnerSide, 1, 2) != 0 {
			continue
		}
		if strings.TrimSpace(item.Result) != "" ||
			strings.TrimSpace(item.Method) != "" ||
			item.Round > 0 || item.TimeSec > 0 {
			continue
		}
		return false
	}
	return true
}

func (m *UFCLiveMonitor) randomInterval() time.Duration {
	if m.maxPollInterval <= m.minPollInterval {
		return m.minPollInterval
	}
	delta := m.maxPollInterval - m.minPollInterval
	return m.minPollInterval + time.Duration(m.random.Int63n(int64(delta)+1))
}

func (m *UFCLiveMonitor) applyBackoff(eventID int64, now time.Time) {
	idx := m.failure[eventID]
	if idx >= len(m.retryBackoffPlan) {
		idx = len(m.retryBackoffPlan) - 1
	}
	m.failure[eventID]++
	m.nextCheckAt[eventID] = now.Add(m.retryBackoffPlan[idx])
}

func (m *UFCLiveMonitor) invalidateCache(ctx context.Context, eventID int64) {
	if m.cache == nil {
		return
	}
	_ = m.cache.InvalidateEvent(ctx, eventID)
	_ = m.cache.InvalidateEvents(ctx)
}
