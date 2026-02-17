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
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, q string) ([]Profile, error) {
	return s.repo.SearchByName(ctx, q)
}

func (s *Service) Get(ctx context.Context, fighterID int64) (Profile, error) {
	return s.repo.GetByID(ctx, fighterID)
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
