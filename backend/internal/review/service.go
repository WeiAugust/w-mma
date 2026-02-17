package review

import (
	"context"
	"time"
)

// PendingArticle represents one article awaiting moderation.
type PendingArticle struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	SourceURL string `json:"source_url"`
}

// Repository defines persistence for review flow.
type Repository interface {
	GetPending(ctx context.Context, pendingID int64) (PendingArticle, error)
	PublishArticle(ctx context.Context, rec PendingArticle) error
	MarkApproved(ctx context.Context, pendingID int64, reviewerID int64) error
	ListPending(ctx context.Context) ([]PendingArticle, error)
}

type Service struct {
	repo  Repository
	cache ArticlesCache
}

type ArticlesCache interface {
	InvalidateArticlesList(ctx context.Context) error
}

func NewService(repo Repository, cache ...ArticlesCache) *Service {
	s := &Service{repo: repo}
	if len(cache) > 0 {
		s.cache = cache[0]
	}
	return s
}

func (s *Service) Approve(ctx context.Context, pendingID int64, reviewerID int64) error {
	rec, err := s.repo.GetPending(ctx, pendingID)
	if err != nil {
		return err
	}
	if err := s.repo.PublishArticle(ctx, rec); err != nil {
		return err
	}
	if err := s.repo.MarkApproved(ctx, pendingID, reviewerID); err != nil {
		return err
	}

	s.invalidateArticlesCache(ctx)
	return nil
}

func (s *Service) ListPending(ctx context.Context) ([]PendingArticle, error) {
	return s.repo.ListPending(ctx)
}

func (s *Service) invalidateArticlesCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.InvalidateArticlesList(ctx)

	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = s.cache.InvalidateArticlesList(context.Background())
	}()
}
