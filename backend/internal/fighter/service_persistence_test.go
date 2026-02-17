package fighter

import (
	"context"
	"testing"
)

type fakeFighterRepo struct {
	searchCalls int
	getCalls    int
}

func (r *fakeFighterRepo) SearchByName(context.Context, string) ([]Profile, error) {
	r.searchCalls++
	return []Profile{{ID: 20, Name: "Alex Pereira"}}, nil
}

func (r *fakeFighterRepo) GetByID(context.Context, int64) (Profile, error) {
	r.getCalls++
	return Profile{ID: 20, Name: "Alex Pereira"}, nil
}

type fakeFighterCacheMiss struct {
	searchSetCalled bool
}

func (c *fakeFighterCacheMiss) GetSearch(context.Context, string) ([]Profile, bool, error) {
	return nil, false, nil
}

func (c *fakeFighterCacheMiss) SetSearch(context.Context, string, []Profile) error {
	c.searchSetCalled = true
	return nil
}

func (c *fakeFighterCacheMiss) GetProfile(context.Context, int64) (Profile, bool, error) {
	return Profile{}, false, nil
}

func (c *fakeFighterCacheMiss) SetProfile(context.Context, int64, Profile) error {
	return nil
}

func TestSearch_UsesCacheAside(t *testing.T) {
	cache := &fakeFighterCacheMiss{}
	repo := &fakeFighterRepo{}
	svc := NewService(repo, cache)
	_, _ = svc.Search(context.Background(), "Alex")
	if repo.searchCalls != 1 {
		t.Fatalf("expected db search")
	}
	if !cache.searchSetCalled {
		t.Fatalf("expected cache fill")
	}
}
