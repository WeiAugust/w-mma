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

func (r *fakeFighterRepo) CreateManual(context.Context, CreateManualInput) (Profile, error) {
	return Profile{ID: 99, Name: "Manual Fighter"}, nil
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

func TestSearch_MatchesNicknameAndChineseName(t *testing.T) {
	svc := NewService(NewInMemoryRepository())

	byNickname, err := svc.Search(context.Background(), "poatan")
	if err != nil {
		t.Fatalf("search by nickname: %v", err)
	}
	if len(byNickname) == 0 || byNickname[0].Name != "Alex Pereira" {
		t.Fatalf("expected Alex Pereira by nickname, got %+v", byNickname)
	}

	byChineseName, err := svc.Search(context.Background(), "安卡拉耶夫")
	if err != nil {
		t.Fatalf("search by chinese name: %v", err)
	}
	if len(byChineseName) == 0 || byChineseName[0].Name != "Magomed Ankalaev" {
		t.Fatalf("expected Magomed Ankalaev by chinese name, got %+v", byChineseName)
	}
}
