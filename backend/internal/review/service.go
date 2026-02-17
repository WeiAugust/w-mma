package review

import (
	"context"
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
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Approve(ctx context.Context, pendingID int64, reviewerID int64) error {
	rec, err := s.repo.GetPending(ctx, pendingID)
	if err != nil {
		return err
	}
	if err := s.repo.PublishArticle(ctx, rec); err != nil {
		return err
	}
	return s.repo.MarkApproved(ctx, pendingID, reviewerID)
}

func (s *Service) ListPending(ctx context.Context) ([]PendingArticle, error) {
	return s.repo.ListPending(ctx)
}
