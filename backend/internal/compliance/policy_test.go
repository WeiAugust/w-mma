package compliance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

type fakeSourceGetter struct {
	item source.DataSource
	err  error
}

func (f *fakeSourceGetter) Get(context.Context, int64) (source.DataSource, error) {
	if f.err != nil {
		return source.DataSource{}, f.err
	}
	return f.item, nil
}

func TestCanPlay_DeniesWhenMissingRights(t *testing.T) {
	policy := NewPlaybackPolicy(&fakeSourceGetter{
		item: source.DataSource{
			RightsDisplay:  true,
			RightsPlayback: false,
		},
	})
	if policy.CanPlay(context.Background(), 1) {
		t.Fatalf("expected playback denied")
	}
}

func TestCanPlay_DeniesWhenExpired(t *testing.T) {
	expiredAt := time.Now().Add(-1 * time.Hour)
	policy := NewPlaybackPolicy(&fakeSourceGetter{
		item: source.DataSource{
			RightsDisplay:   true,
			RightsPlayback:  true,
			RightsExpiresAt: &expiredAt,
		},
	})
	if policy.CanPlay(context.Background(), 1) {
		t.Fatalf("expected playback denied")
	}
}

func TestCanPlay_AllowsWhenValid(t *testing.T) {
	expiresAt := time.Now().Add(1 * time.Hour)
	policy := NewPlaybackPolicy(&fakeSourceGetter{
		item: source.DataSource{
			RightsDisplay:   true,
			RightsPlayback:  true,
			RightsExpiresAt: &expiresAt,
		},
	})
	if !policy.CanPlay(context.Background(), 1) {
		t.Fatalf("expected playback allowed")
	}
}

func TestCanPlay_DeniesWhenLookupFails(t *testing.T) {
	policy := NewPlaybackPolicy(&fakeSourceGetter{err: errors.New("boom")})
	if policy.CanPlay(context.Background(), 1) {
		t.Fatalf("expected playback denied")
	}
}
