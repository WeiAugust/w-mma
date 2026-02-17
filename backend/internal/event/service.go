package event

import (
	"context"
	"errors"
	"sync"
)

// Bout is one matchup in an event card.
type Bout struct {
	ID            int64  `json:"id"`
	RedFighterID  int64  `json:"red_fighter_id"`
	BlueFighterID int64  `json:"blue_fighter_id"`
	Result        string `json:"result"`
	WinnerID      int64  `json:"winner_id"`
}

// Card is event detail with all bouts.
type Card struct {
	ID            int64  `json:"id"`
	Org           string `json:"org"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	PosterURL     string `json:"poster_url,omitempty"`
	PromoVideoURL string `json:"promo_video_url,omitempty"`
	Bouts         []Bout `json:"bouts"`
}

// EventSummary is list item for schedule page.
type EventSummary struct {
	ID            int64  `json:"id"`
	Org           string `json:"org"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	StartsAt      string `json:"starts_at"`
	PosterURL     string `json:"poster_url,omitempty"`
	PromoVideoURL string `json:"promo_video_url,omitempty"`
}

type UpdateEventInput struct {
	Name   string
	Status string
}

type Repository interface {
	GetEventCard(ctx context.Context, eventID int64) (Card, error)
	ListEvents(ctx context.Context) ([]EventSummary, error)
	UpdateEvent(ctx context.Context, eventID int64, input UpdateEventInput) error
}

type Service struct {
	repo  Repository
	cache EventCache
}

type EventCache interface {
	GetEventCard(ctx context.Context, eventID int64) (Card, bool, error)
	SetEventCard(ctx context.Context, eventID int64, card Card, status string) error
	GetEvents(ctx context.Context) ([]EventSummary, bool, error)
	SetEvents(ctx context.Context, events []EventSummary) error
	InvalidateEvent(ctx context.Context, eventID int64) error
	InvalidateEvents(ctx context.Context) error
}

func NewService(repo Repository, cache ...EventCache) *Service {
	s := &Service{repo: repo}
	if len(cache) > 0 {
		s.cache = cache[0]
	}
	return s
}

func (s *Service) GetEventCard(ctx context.Context, eventID int64) (Card, error) {
	if s.cache != nil {
		if card, ok, err := s.cache.GetEventCard(ctx, eventID); err == nil && ok {
			return card, nil
		}
	}

	card, err := s.repo.GetEventCard(ctx, eventID)
	if err != nil {
		return Card{}, err
	}
	if s.cache != nil {
		_ = s.cache.SetEventCard(ctx, eventID, card, card.Status)
	}
	return card, nil
}

func (s *Service) ListEvents(ctx context.Context) ([]EventSummary, error) {
	if s.cache != nil {
		if items, ok, err := s.cache.GetEvents(ctx); err == nil && ok {
			return items, nil
		}
	}

	items, err := s.repo.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetEvents(ctx, items)
	}
	return items, nil
}

func (s *Service) UpdateEvent(ctx context.Context, eventID int64, input UpdateEventInput) error {
	if err := s.repo.UpdateEvent(ctx, eventID, input); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.InvalidateEvent(ctx, eventID)
		_ = s.cache.InvalidateEvents(ctx)
	}
	return nil
}

type InMemoryRepository struct {
	mu sync.Mutex

	cards  map[int64]Card
	events []EventSummary
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		cards: map[int64]Card{
			10: {
				ID:     10,
				Org:    "UFC",
				Name:   "UFC Fight Night 10",
				Status: "live",
				Bouts: []Bout{
					{ID: 1001, RedFighterID: 20, BlueFighterID: 21, Result: "pending", WinnerID: 0},
				},
			},
		},
		events: []EventSummary{
			{ID: 10, Org: "UFC", Name: "UFC Fight Night 10", Status: "live", StartsAt: "2026-02-17T20:00:00Z"},
		},
	}
}

func (r *InMemoryRepository) GetEventCard(_ context.Context, eventID int64) (Card, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	card, ok := r.cards[eventID]
	if !ok {
		return Card{}, errors.New("event not found")
	}
	return card, nil
}

func (r *InMemoryRepository) ListEvents(context.Context) ([]EventSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]EventSummary, len(r.events))
	copy(items, r.events)
	return items, nil
}

func (r *InMemoryRepository) UpdateEvent(_ context.Context, eventID int64, input UpdateEventInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	card, ok := r.cards[eventID]
	if !ok {
		return errors.New("event not found")
	}
	if input.Name != "" {
		card.Name = input.Name
	}
	if input.Status != "" {
		card.Status = input.Status
	}
	r.cards[eventID] = card

	for i := range r.events {
		if r.events[i].ID != eventID {
			continue
		}
		if input.Name != "" {
			r.events[i].Name = input.Name
		}
		if input.Status != "" {
			r.events[i].Status = input.Status
		}
		break
	}

	return nil
}

// UpsertBoutResult updates one bout result in-memory for live-update flow.
func (r *InMemoryRepository) UpsertBoutResult(_ context.Context, eventID int64, boutID int64, winnerID int64, result string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	card, ok := r.cards[eventID]
	if !ok {
		return errors.New("event not found")
	}

	for i := range card.Bouts {
		if card.Bouts[i].ID != boutID {
			continue
		}
		card.Bouts[i].WinnerID = winnerID
		card.Bouts[i].Result = result
		r.cards[eventID] = card
		return nil
	}

	return errors.New("bout not found")
}
