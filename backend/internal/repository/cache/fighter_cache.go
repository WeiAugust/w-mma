package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/bajiaozhi/w-mma/backend/internal/fighter"
)

type FighterCache struct {
	client redis.Cmdable
}

func NewFighterCache(client redis.Cmdable) *FighterCache {
	return &FighterCache{client: client}
}

func fighterProfileKey(id int64) string {
	return fmt.Sprintf("cache:fighter:detail:v1:%d", id)
}

func fighterSearchKey(q string) string {
	return fmt.Sprintf("cache:fighter:search:v1:%s", q)
}

func (c *FighterCache) GetSearch(ctx context.Context, q string) ([]fighter.Profile, bool, error) {
	payload, err := c.client.Get(ctx, fighterSearchKey(q)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var items []fighter.Profile
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, false, err
	}
	return items, true, nil
}

func (c *FighterCache) SetSearch(ctx context.Context, q string, items []fighter.Profile) error {
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fighterSearchKey(q), payload, 120*time.Second).Err()
}

func (c *FighterCache) GetProfile(ctx context.Context, fighterID int64) (fighter.Profile, bool, error) {
	payload, err := c.client.Get(ctx, fighterProfileKey(fighterID)).Result()
	if err == redis.Nil {
		return fighter.Profile{}, false, nil
	}
	if err != nil {
		return fighter.Profile{}, false, err
	}
	var profile fighter.Profile
	if err := json.Unmarshal([]byte(payload), &profile); err != nil {
		return fighter.Profile{}, false, err
	}
	return profile, true, nil
}

func (c *FighterCache) SetProfile(ctx context.Context, fighterID int64, profile fighter.Profile) error {
	payload, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fighterProfileKey(fighterID), payload, 300*time.Second).Err()
}
