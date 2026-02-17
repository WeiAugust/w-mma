package takedown

import (
	"context"
	"errors"
	"sync"
)

const (
	ActionOfflined = "offlined"
	ActionRejected = "rejected"
)

var (
	ErrUnsupportedTargetType = errors.New("unsupported target type")
	ErrTicketNotFound        = errors.New("takedown ticket not found")
)

type Ticket struct {
	ID          int64  `json:"id"`
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	Reason      string `json:"reason"`
	Complainant string `json:"complainant,omitempty"`
	EvidenceURL string `json:"evidence_url,omitempty"`
	Status      string `json:"status"`
	Action      string `json:"action,omitempty"`
}

type CreateInput struct {
	TargetType  string
	TargetID    int64
	Reason      string
	Complainant string
	EvidenceURL string
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Ticket, error)
	Get(ctx context.Context, ticketID int64) (Ticket, error)
	Resolve(ctx context.Context, ticketID int64, action string) error
}

type ContentOffliner interface {
	OfflineArticle(ctx context.Context, articleID int64) error
}

type CacheInvalidator interface {
	InvalidateArticlesList(ctx context.Context) error
}

type Service struct {
	repo     Repository
	offliner ContentOffliner
	cache    CacheInvalidator
}

func NewService(repo Repository, offliner ContentOffliner, cache CacheInvalidator) *Service {
	return &Service{
		repo:     repo,
		offliner: offliner,
		cache:    cache,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Ticket, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Resolve(ctx context.Context, ticketID int64, action string) error {
	ticket, err := s.repo.Get(ctx, ticketID)
	if err != nil {
		return err
	}

	if action == ActionOfflined {
		switch ticket.TargetType {
		case "article":
			if err := s.offliner.OfflineArticle(ctx, ticket.TargetID); err != nil {
				return err
			}
			if s.cache != nil {
				_ = s.cache.InvalidateArticlesList(ctx)
			}
		default:
			return ErrUnsupportedTargetType
		}
	}

	return s.repo.Resolve(ctx, ticketID, action)
}

type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]Ticket
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  map[int64]Ticket{},
	}
}

func (r *InMemoryRepository) Create(_ context.Context, input CreateInput) (Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item := Ticket{
		ID:          r.nextID,
		TargetType:  input.TargetType,
		TargetID:    input.TargetID,
		Reason:      input.Reason,
		Complainant: input.Complainant,
		EvidenceURL: input.EvidenceURL,
		Status:      "open",
	}
	r.items[item.ID] = item
	r.nextID++
	return item, nil
}

func (r *InMemoryRepository) Get(_ context.Context, ticketID int64) (Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[ticketID]
	if !ok {
		return Ticket{}, ErrTicketNotFound
	}
	return item, nil
}

func (r *InMemoryRepository) Resolve(_ context.Context, ticketID int64, action string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[ticketID]
	if !ok {
		return ErrTicketNotFound
	}
	item.Status = "resolved"
	item.Action = action
	r.items[ticketID] = item
	return nil
}
