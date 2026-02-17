package event

import (
	"context"
	"errors"
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
	ID     int64  `json:"id"`
	Org    string `json:"org"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Bouts  []Bout `json:"bouts"`
}

// EventSummary is list item for schedule page.
type EventSummary struct {
	ID       int64  `json:"id"`
	Org      string `json:"org"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	StartsAt string `json:"starts_at"`
}

type Repository interface {
	GetEventCard(ctx context.Context, eventID int64) (Card, error)
	ListEvents(ctx context.Context) ([]EventSummary, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetEventCard(ctx context.Context, eventID int64) (Card, error) {
	return s.repo.GetEventCard(ctx, eventID)
}

func (s *Service) ListEvents(ctx context.Context) ([]EventSummary, error) {
	return s.repo.ListEvents(ctx)
}

type InMemoryRepository struct {
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
	card, ok := r.cards[eventID]
	if !ok {
		return Card{}, errors.New("event not found")
	}
	return card, nil
}

func (r *InMemoryRepository) ListEvents(context.Context) ([]EventSummary, error) {
	return r.events, nil
}
