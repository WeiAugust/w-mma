package source

import (
	"context"
	"errors"
	"sort"
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
	IsBuiltin       bool       `json:"is_builtin"`
	RightsDisplay   bool       `json:"rights_display"`
	RightsPlayback  bool       `json:"rights_playback"`
	RightsAISummary bool       `json:"rights_ai_summary"`
	RightsExpiresAt *time.Time `json:"rights_expires_at,omitempty"`
	RightsProofURL  string     `json:"rights_proof_url,omitempty"`
	LastFetchAt     *time.Time `json:"last_fetch_at,omitempty"`
	LastFetchStatus string     `json:"last_fetch_status,omitempty"`
	LastFetchError  string     `json:"last_fetch_error,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type CreateInput struct {
	Name            string
	SourceType      string
	Platform        string
	AccountID       string
	SourceURL       string
	ParserKind      string
	Enabled         bool
	IsBuiltin       bool
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

type ListFilter struct {
	IncludeDeleted bool
	SourceType     string
	Platform       string
	Enabled        *bool
	IsBuiltin      *bool
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (DataSource, error)
	List(ctx context.Context, filter ListFilter) ([]DataSource, error)
	Get(ctx context.Context, sourceID int64, includeDeleted bool) (DataSource, error)
	Update(ctx context.Context, sourceID int64, input UpdateInput) error
	SetEnabled(ctx context.Context, sourceID int64, enabled bool) error
	SoftDelete(ctx context.Context, sourceID int64) error
	Restore(ctx context.Context, sourceID int64) error
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

func (s *Service) List(ctx context.Context, filter ListFilter) ([]DataSource, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, sourceID int64) (DataSource, error) {
	return s.repo.Get(ctx, sourceID, false)
}

func (s *Service) GetAny(ctx context.Context, sourceID int64) (DataSource, error) {
	return s.repo.Get(ctx, sourceID, true)
}

func (s *Service) Update(ctx context.Context, sourceID int64, input UpdateInput) error {
	return s.repo.Update(ctx, sourceID, input)
}

func (s *Service) Toggle(ctx context.Context, sourceID int64) error {
	item, err := s.repo.Get(ctx, sourceID, false)
	if err != nil {
		return err
	}
	return s.repo.SetEnabled(ctx, sourceID, !item.Enabled)
}

func (s *Service) Delete(ctx context.Context, sourceID int64) error {
	return s.repo.SoftDelete(ctx, sourceID)
}

func (s *Service) Restore(ctx context.Context, sourceID int64) error {
	return s.repo.Restore(ctx, sourceID)
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
		IsBuiltin:       input.IsBuiltin,
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

func (r *InMemoryRepository) List(_ context.Context, filter ListFilter) ([]DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]int64, 0, len(r.items))
	for id := range r.items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	items := make([]DataSource, 0, len(ids))
	for _, id := range ids {
		item := r.items[id]
		if !matchFilter(item, filter) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *InMemoryRepository) Get(_ context.Context, sourceID int64, includeDeleted bool) (DataSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok {
		return DataSource{}, ErrSourceNotFound
	}
	if !includeDeleted && item.DeletedAt != nil {
		return DataSource{}, ErrSourceNotFound
	}
	return item, nil
}

func (r *InMemoryRepository) Update(_ context.Context, sourceID int64, input UpdateInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok || item.DeletedAt != nil {
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
	if !ok || item.DeletedAt != nil {
		return ErrSourceNotFound
	}
	item.Enabled = enabled
	r.items[sourceID] = item
	return nil
}

func (r *InMemoryRepository) SoftDelete(_ context.Context, sourceID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok || item.DeletedAt != nil {
		return ErrSourceNotFound
	}
	now := time.Now()
	item.DeletedAt = &now
	r.items[sourceID] = item
	return nil
}

func (r *InMemoryRepository) Restore(_ context.Context, sourceID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[sourceID]
	if !ok {
		return ErrSourceNotFound
	}
	item.DeletedAt = nil
	r.items[sourceID] = item
	return nil
}

func matchFilter(item DataSource, filter ListFilter) bool {
	if !filter.IncludeDeleted && item.DeletedAt != nil {
		return false
	}
	if filter.SourceType != "" && item.SourceType != filter.SourceType {
		return false
	}
	if filter.Platform != "" && item.Platform != filter.Platform {
		return false
	}
	if filter.Enabled != nil && item.Enabled != *filter.Enabled {
		return false
	}
	if filter.IsBuiltin != nil && item.IsBuiltin != *filter.IsBuiltin {
		return false
	}
	return true
}
