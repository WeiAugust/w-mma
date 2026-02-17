package source

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidSourceType = errors.New("invalid source type")
	ErrSourceNotFound    = errors.New("source not found")
)

type DataSource struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	SourceType      string     `json:"source_type"`
	Platform        string     `json:"platform"`
	AccountID       string     `json:"account_id,omitempty"`
	SourceURL       string     `json:"source_url"`
	ParserKind      string     `json:"parser_kind"`
	Enabled         bool       `json:"enabled"`
	RightsDisplay   bool       `json:"rights_display"`
	RightsPlayback  bool       `json:"rights_playback"`
	RightsAISummary bool       `json:"rights_ai_summary"`
	RightsExpiresAt *time.Time `json:"rights_expires_at,omitempty"`
	RightsProofURL  string     `json:"rights_proof_url,omitempty"`
}

type CreateInput struct {
	Name            string
	SourceType      string
	Platform        string
	AccountID       string
	SourceURL       string
	ParserKind      string
	Enabled         bool
	RightsDisplay   bool
	RightsPlayback  bool
	RightsAISummary bool
	RightsExpiresAt time.Time
	RightsProofURL  string
}

type UpdateInput struct {
	Name            string
	Platform        string
	AccountID       *string
	SourceURL       string
	ParserKind      string
	RightsDisplay   *bool
	RightsPlayback  *bool
	RightsAISummary *bool
	RightsExpiresAt *time.Time
	RightsProofURL  *string
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (DataSource, error)
	List(ctx context.Context) ([]DataSource, error)
	Get(ctx context.Context, sourceID int64) (DataSource, error)
	Update(ctx context.Context, sourceID int64, input UpdateInput) error
	SetEnabled(ctx context.Context, sourceID int64, enabled bool) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (DataSource, error) {
	if !isValidSourceType(input.SourceType) {
		return DataSource{}, ErrInvalidSourceType
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) List(ctx context.Context) ([]DataSource, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, sourceID int64) (DataSource, error) {
	return s.repo.Get(ctx, sourceID)
}

func (s *Service) Update(ctx context.Context, sourceID int64, input UpdateInput) error {
	return s.repo.Update(ctx, sourceID, input)
}

func (s *Service) Toggle(ctx context.Context, sourceID int64) error {
	item, err := s.repo.Get(ctx, sourceID)
	if err != nil {
		return err
	}
	return s.repo.SetEnabled(ctx, sourceID, !item.Enabled)
}

func isValidSourceType(sourceType string) bool {
	switch sourceType {
	case "news", "schedule", "fighter":
		return true
	default:
		return false
	}
}

type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]DataSource
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  make(map[int64]DataSource),
	}
}

func (r *InMemoryRepository) Create(_ context.Context, input CreateInput) (DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item := DataSource{
		ID:              r.nextID,
		Name:            input.Name,
		SourceType:      input.SourceType,
		Platform:        input.Platform,
		AccountID:       input.AccountID,
		SourceURL:       input.SourceURL,
		ParserKind:      input.ParserKind,
		Enabled:         input.Enabled,
		RightsDisplay:   input.RightsDisplay,
		RightsPlayback:  input.RightsPlayback,
		RightsAISummary: input.RightsAISummary,
	}
	if !input.RightsExpiresAt.IsZero() {
		expires := input.RightsExpiresAt
		item.RightsExpiresAt = &expires
	}
	item.RightsProofURL = input.RightsProofURL

	r.items[item.ID] = item
	r.nextID++
	return item, nil
}

func (r *InMemoryRepository) List(_ context.Context) ([]DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]DataSource, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items, nil
}

func (r *InMemoryRepository) Get(_ context.Context, sourceID int64) (DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok {
		return DataSource{}, ErrSourceNotFound
	}
	return item, nil
}

func (r *InMemoryRepository) Update(_ context.Context, sourceID int64, input UpdateInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok {
		return ErrSourceNotFound
	}

	if input.Name != "" {
		item.Name = input.Name
	}
	if input.Platform != "" {
		item.Platform = input.Platform
	}
	if input.AccountID != nil {
		item.AccountID = *input.AccountID
	}
	if input.SourceURL != "" {
		item.SourceURL = input.SourceURL
	}
	if input.ParserKind != "" {
		item.ParserKind = input.ParserKind
	}
	if input.RightsDisplay != nil {
		item.RightsDisplay = *input.RightsDisplay
	}
	if input.RightsPlayback != nil {
		item.RightsPlayback = *input.RightsPlayback
	}
	if input.RightsAISummary != nil {
		item.RightsAISummary = *input.RightsAISummary
	}
	if input.RightsExpiresAt != nil {
		item.RightsExpiresAt = input.RightsExpiresAt
	}
	if input.RightsProofURL != nil {
		item.RightsProofURL = *input.RightsProofURL
	}

	r.items[sourceID] = item
	return nil
}

func (r *InMemoryRepository) SetEnabled(_ context.Context, sourceID int64, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok {
		return ErrSourceNotFound
	}
	item.Enabled = enabled
	r.items[sourceID] = item
	return nil
}
