package review

import (
	"context"
	"testing"
)

type fakeReviewRepo struct {
	pending map[int64]PendingArticle

	articlePublished bool
	approved         bool
	approvedBy       int64
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{
		pending: map[int64]PendingArticle{
			101: {
				ID:        101,
				Title:     "news-a",
				Summary:   "summary-a",
				SourceURL: "https://example.com/a",
			},
		},
	}
}

func (r *fakeReviewRepo) GetPending(_ context.Context, pendingID int64) (PendingArticle, error) {
	return r.pending[pendingID], nil
}

func (r *fakeReviewRepo) PublishArticle(context.Context, PendingArticle) error {
	r.articlePublished = true
	return nil
}

func (r *fakeReviewRepo) MarkApproved(_ context.Context, _ int64, reviewerID int64) error {
	r.approved = true
	r.approvedBy = reviewerID
	return nil
}

func (r *fakeReviewRepo) ListPending(context.Context) ([]PendingArticle, error) {
	items := make([]PendingArticle, 0, len(r.pending))
	for _, p := range r.pending {
		items = append(items, p)
	}
	return items, nil
}

func TestApprove_PublishesArticle(t *testing.T) {
	repo := newFakeReviewRepo()
	svc := NewService(repo)
	err := svc.Approve(context.Background(), 101, 9001)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.articlePublished {
		t.Fatalf("expected article to be published")
	}
}
