package media

import (
	"context"
	"errors"
	"sort"
	"sync"
)

var (
	ErrInvalidOwnerType = errors.New("invalid owner type")
	ErrInvalidMediaType = errors.New("invalid media type")
)

type Asset struct {
	ID        int64  `json:"id"`
	OwnerType string `json:"owner_type"`
	OwnerID   int64  `json:"owner_id"`
	MediaType string `json:"media_type"`
	URL       string `json:"url"`
	CoverURL  string `json:"cover_url,omitempty"`
	Title     string `json:"title,omitempty"`
	SortNo    int    `json:"sort_no"`
}

type AttachInput struct {
	OwnerType string
	OwnerID   int64
	MediaType string
	URL       string
	CoverURL  string
	Title     string
	SortNo    int
}

type Repository interface {
	Attach(ctx context.Context, input AttachInput) (Asset, error)
	ListByOwner(ctx context.Context, ownerType string, ownerID int64) ([]Asset, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Attach(ctx context.Context, input AttachInput) (Asset, error) {
	if !isValidOwnerType(input.OwnerType) {
		return Asset{}, ErrInvalidOwnerType
	}
	if !isValidMediaType(input.MediaType) {
		return Asset{}, ErrInvalidMediaType
	}
	return s.repo.Attach(ctx, input)
}

func (s *Service) ListByOwner(ctx context.Context, ownerType string, ownerID int64) ([]Asset, error) {
	return s.repo.ListByOwner(ctx, ownerType, ownerID)
}

func isValidOwnerType(ownerType string) bool {
	switch ownerType {
	case "article", "event", "fighter":
		return true
	default:
		return false
	}
}

func isValidMediaType(mediaType string) bool {
	switch mediaType {
	case "image", "video":
		return true
	default:
		return false
	}
}

type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int64
	items  []Asset
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  []Asset{},
	}
}

func (r *InMemoryRepository) Attach(_ context.Context, input AttachInput) (Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	asset := Asset{
		ID:        r.nextID,
		OwnerType: input.OwnerType,
		OwnerID:   input.OwnerID,
		MediaType: input.MediaType,
		URL:       input.URL,
		CoverURL:  input.CoverURL,
		Title:     input.Title,
		SortNo:    input.SortNo,
	}
	r.items = append(r.items, asset)
	r.nextID++
	return asset, nil
}

func (r *InMemoryRepository) ListByOwner(_ context.Context, ownerType string, ownerID int64) ([]Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Asset, 0)
	for _, item := range r.items {
		if item.OwnerType == ownerType && item.OwnerID == ownerID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].SortNo < items[j].SortNo
	})
	return items, nil
}
