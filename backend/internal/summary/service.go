package summary

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	StatusPending        = "pending"
	StatusRunning        = "running"
	StatusDone           = "done"
	StatusFailed         = "failed"
	StatusManualRequired = "manual_required"
)

var ErrJobNotFound = errors.New("summary job not found")

type Config struct {
	Provider string
	APIBase  string
	APIKey   string
}

type Job struct {
	ID         int64     `json:"id"`
	SourceID   int64     `json:"source_id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Status     string    `json:"status"`
	Provider   string    `json:"provider,omitempty"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateInput struct {
	SourceID   int64
	TargetType string
	TargetID   int64
	Status     string
	Provider   string
	ErrorMsg   string
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Job, error)
	List(ctx context.Context) ([]Job, error)
	Get(ctx context.Context, jobID int64) (Job, error)
	UpdateStatus(ctx context.Context, jobID int64, status string, errorMsg string) error
}

type Service struct {
	repo   Repository
	config Config
}

func NewService(repo Repository, config Config) *Service {
	return &Service{
		repo:   repo,
		config: config,
	}
}

func (s *Service) CreateArticleJob(ctx context.Context, sourceID int64, articleID int64) (Job, error) {
	status := StatusPending
	errMsg := ""
	if s.config.APIKey == "" {
		status = StatusManualRequired
		errMsg = "api key missing, fallback to manual"
	}
	return s.repo.Create(ctx, CreateInput{
		SourceID:   sourceID,
		TargetType: "article",
		TargetID:   articleID,
		Status:     status,
		Provider:   s.config.Provider,
		ErrorMsg:   errMsg,
	})
}

func (s *Service) ListJobs(ctx context.Context) ([]Job, error) {
	return s.repo.List(ctx)
}

type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]Job
	order  []int64
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  map[int64]Job{},
		order:  []int64{},
	}
}

func (r *InMemoryRepository) Create(_ context.Context, input CreateInput) (Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	item := Job{
		ID:         r.nextID,
		SourceID:   input.SourceID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Status:     input.Status,
		Provider:   input.Provider,
		ErrorMsg:   input.ErrorMsg,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.items[item.ID] = item
	r.order = append(r.order, item.ID)
	r.nextID++
	return item, nil
}

func (r *InMemoryRepository) List(_ context.Context) ([]Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Job, 0, len(r.order))
	for _, id := range r.order {
		items = append(items, r.items[id])
	}
	return items, nil
}

func (r *InMemoryRepository) Get(_ context.Context, jobID int64) (Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[jobID]
	if !ok {
		return Job{}, ErrJobNotFound
	}
	return item, nil
}

func (r *InMemoryRepository) UpdateStatus(_ context.Context, jobID int64, status string, errorMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[jobID]
	if !ok {
		return ErrJobNotFound
	}
	item.Status = status
	item.ErrorMsg = errorMsg
	item.UpdatedAt = time.Now()
	r.items[jobID] = item
	return nil
}
