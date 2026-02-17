package fighter

import (
	"context"
	"errors"
	"strings"
)

// Profile contains fighter details and recent updates.
type Profile struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Country string   `json:"country"`
	Record  string   `json:"record"`
	Updates []string `json:"updates"`
}

type Repository interface {
	SearchByName(ctx context.Context, q string) ([]Profile, error)
	GetByID(ctx context.Context, fighterID int64) (Profile, error)
}

type Service struct {
	repo  Repository
	cache FighterCache
}

type FighterCache interface {
	GetSearch(ctx context.Context, q string) ([]Profile, bool, error)
	SetSearch(ctx context.Context, q string, items []Profile) error
	GetProfile(ctx context.Context, fighterID int64) (Profile, bool, error)
	SetProfile(ctx context.Context, fighterID int64, profile Profile) error
}

func NewService(repo Repository, cache ...FighterCache) *Service {
	s := &Service{repo: repo}
	if len(cache) > 0 {
		s.cache = cache[0]
	}
	return s
}

func (s *Service) Search(ctx context.Context, q string) ([]Profile, error) {
	if s.cache != nil {
		if items, ok, err := s.cache.GetSearch(ctx, q); err == nil && ok {
			return items, nil
		}
	}

	items, err := s.repo.SearchByName(ctx, q)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSearch(ctx, q, items)
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, fighterID int64) (Profile, error) {
	if s.cache != nil {
		if profile, ok, err := s.cache.GetProfile(ctx, fighterID); err == nil && ok {
			return profile, nil
		}
	}

	profile, err := s.repo.GetByID(ctx, fighterID)
	if err != nil {
		return Profile{}, err
	}
	if s.cache != nil {
		_ = s.cache.SetProfile(ctx, fighterID, profile)
	}
	return profile, nil
}

type InMemoryRepository struct {
	fighters map[int64]Profile
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{fighters: map[int64]Profile{
		20: {
			ID:      20,
			Name:    "Alex Pereira",
			Country: "Brazil",
			Record:  "10-2",
			Updates: []string{"Win vs Jan", "Title defense confirmed"},
		},
		21: {
			ID:      21,
			Name:    "Magomed Ankalaev",
			Country: "Russia",
			Record:  "19-1-1",
			Updates: []string{"Camp started", "Media day completed"},
		},
	}}
}

func (r *InMemoryRepository) SearchByName(_ context.Context, q string) ([]Profile, error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return []Profile{}, nil
	}
	res := make([]Profile, 0)
	for _, p := range r.fighters {
		if strings.Contains(strings.ToLower(p.Name), q) {
			res = append(res, p)
		}
	}
	return res, nil
}

func (r *InMemoryRepository) GetByID(_ context.Context, fighterID int64) (Profile, error) {
	p, ok := r.fighters[fighterID]
	if !ok {
		return Profile{}, errors.New("fighter not found")
	}
	return p, nil
}
