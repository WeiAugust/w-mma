package compliance

import (
	"context"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

type sourceGetter interface {
	Get(ctx context.Context, sourceID int64) (source.DataSource, error)
}

type PlaybackPolicy struct {
	getter sourceGetter
	now    func() time.Time
}

func NewPlaybackPolicy(getter sourceGetter) *PlaybackPolicy {
	return &PlaybackPolicy{
		getter: getter,
		now:    time.Now,
	}
}

func (p *PlaybackPolicy) CanPlay(ctx context.Context, sourceID int64) bool {
	if p == nil || p.getter == nil || sourceID <= 0 {
		return false
	}

	item, err := p.getter.Get(ctx, sourceID)
	if err != nil {
		return false
	}
	if !item.RightsDisplay || !item.RightsPlayback {
		return false
	}
	if item.RightsExpiresAt != nil && p.now().After(*item.RightsExpiresAt) {
		return false
	}
	return true
}
