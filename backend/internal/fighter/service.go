package fighter

import (
	"context"
	"errors"
	"strings"
)

// Profile contains fighter details and recent updates.
type Profile struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	NameZH        string            `json:"name_zh,omitempty"`
	Nickname      string            `json:"nickname,omitempty"`
	Country       string            `json:"country"`
	Record        string            `json:"record"`
	WeightClass   string            `json:"weight_class,omitempty"`
	AvatarURL     string            `json:"avatar_url,omitempty"`
	IntroVideoURL string            `json:"intro_video_url,omitempty"`
	Stats         map[string]string `json:"stats,omitempty"`
	Records       map[string]string `json:"records,omitempty"`
	Updates       []string          `json:"updates"`
}

type CreateManualInput struct {
	SourceID      int64
	Name          string
	NameZH        string
	Nickname      string
	Country       string
	Record        string
	WeightClass   string
	AvatarURL     string
	IntroVideoURL string
}

type Repository interface {
	SearchByName(ctx context.Context, q string) ([]Profile, error)
	GetByID(ctx context.Context, fighterID int64) (Profile, error)
	CreateManual(ctx context.Context, input CreateManualInput) (Profile, error)
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

func (s *Service) CreateManual(ctx context.Context, input CreateManualInput) (Profile, error) {
	return s.repo.CreateManual(ctx, input)
}

type InMemoryRepository struct {
	fighters map[int64]Profile
	nextID   int64
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{fighters: map[int64]Profile{
		20: {
			ID:          20,
			Name:        "Alex Pereira",
			NameZH:      "亚历克斯 佩雷拉",
			Nickname:    "Poatan",
			Country:     "Brazil",
			Record:      "10-2",
			WeightClass: "Light Heavyweight",
			Updates:     []string{"Win vs Jan", "Title defense confirmed"},
		},
		21: {
			ID:          21,
			Name:        "Magomed Ankalaev",
			NameZH:      "马戈梅德 安卡拉耶夫",
			Nickname:    "Ankalaev",
			Country:     "Russia",
			Record:      "19-1-1",
			WeightClass: "Light Heavyweight",
			Updates:     []string{"Camp started", "Media day completed"},
		},
	}, nextID: 22}
}

func (r *InMemoryRepository) SearchByName(_ context.Context, q string) ([]Profile, error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return []Profile{}, nil
	}
	res := make([]Profile, 0)
	for _, p := range r.fighters {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Nickname), q) ||
			strings.Contains(strings.ToLower(p.NameZH), q) {
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

func (r *InMemoryRepository) CreateManual(_ context.Context, input CreateManualInput) (Profile, error) {
	profile := Profile{
		ID:            r.nextID,
		Name:          input.Name,
		NameZH:        input.NameZH,
		Nickname:      input.Nickname,
		Country:       input.Country,
		Record:        input.Record,
		WeightClass:   input.WeightClass,
		AvatarURL:     input.AvatarURL,
		IntroVideoURL: input.IntroVideoURL,
		Updates:       []string{},
	}
	r.fighters[profile.ID] = profile
	r.nextID++
	return profile, nil
}
