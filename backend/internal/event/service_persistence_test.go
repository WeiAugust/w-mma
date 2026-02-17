package event

import (
	"context"
	"testing"
)

type fakeEventRepo struct {
	getCardCalls int
	listCalls    int
}

func (r *fakeEventRepo) GetEventCard(context.Context, int64) (Card, error) {
	r.getCardCalls++
	return Card{ID: 10, Name: "UFC 10", Status: "live"}, nil
}

func (r *fakeEventRepo) ListEvents(context.Context) ([]EventSummary, error) {
	r.listCalls++
	return []EventSummary{{ID: 10, Name: "UFC 10", Org: "UFC"}}, nil
}

func (r *fakeEventRepo) UpdateEvent(context.Context, int64, UpdateEventInput) error {
	return nil
}

type fakeEventCacheMiss struct {
	setCardCalled bool
	setListCalled bool
}

func (c *fakeEventCacheMiss) GetEventCard(context.Context, int64) (Card, bool, error) {
	return Card{}, false, nil
}

func (c *fakeEventCacheMiss) SetEventCard(context.Context, int64, Card, string) error {
	c.setCardCalled = true
	return nil
}

func (c *fakeEventCacheMiss) GetEvents(context.Context) ([]EventSummary, bool, error) {
	return nil, false, nil
}

func (c *fakeEventCacheMiss) SetEvents(context.Context, []EventSummary) error {
	c.setListCalled = true
	return nil
}

func (c *fakeEventCacheMiss) InvalidateEvent(context.Context, int64) error {
	return nil
}

func (c *fakeEventCacheMiss) InvalidateEvents(context.Context) error {
	return nil
}

func TestGetEventCard_ReadsFromCacheThenDB(t *testing.T) {
	cache := &fakeEventCacheMiss{}
	repo := &fakeEventRepo{}
	svc := NewService(repo, cache)
	_, _ = svc.GetEventCard(context.Background(), 10)
	if repo.getCardCalls != 1 {
		t.Fatalf("expected db fallback")
	}
	if !cache.setCardCalled {
		t.Fatalf("expected cache fill")
	}
}
