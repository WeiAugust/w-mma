package takedown

import (
	"context"
	"testing"
)

type fakeRepo struct {
	ticket      Ticket
	resolvedIDs []int64
}

func (f *fakeRepo) Create(context.Context, CreateInput) (Ticket, error) {
	return f.ticket, nil
}

func (f *fakeRepo) Get(context.Context, int64) (Ticket, error) {
	return f.ticket, nil
}

func (f *fakeRepo) Resolve(_ context.Context, ticketID int64, action string) error {
	f.resolvedIDs = append(f.resolvedIDs, ticketID)
	f.ticket.Status = "resolved"
	f.ticket.Action = action
	return nil
}

type fakeOffliner struct {
	offlinedIDs []int64
}

func (f *fakeOffliner) OfflineArticle(_ context.Context, articleID int64) error {
	f.offlinedIDs = append(f.offlinedIDs, articleID)
	return nil
}

type fakeCache struct {
	invalidated bool
}

func (f *fakeCache) InvalidateArticlesList(context.Context) error {
	f.invalidated = true
	return nil
}

func TestResolveTakedown_OfflinesTargetAndInvalidatesCache(t *testing.T) {
	repo := &fakeRepo{
		ticket: Ticket{
			ID:         10,
			TargetType: "article",
			TargetID:   101,
			Status:     "open",
		},
	}
	offliner := &fakeOffliner{}
	cache := &fakeCache{}
	svc := NewService(repo, offliner, cache)

	if err := svc.Resolve(context.Background(), 10, ActionOfflined); err != nil {
		t.Fatalf("resolve takedown: %v", err)
	}
	if len(offliner.offlinedIDs) != 1 || offliner.offlinedIDs[0] != 101 {
		t.Fatalf("expected article 101 to be offlined")
	}
	if !cache.invalidated {
		t.Fatalf("expected article cache invalidated")
	}
}
