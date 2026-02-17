package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/bajiaozhi/w-mma/backend/internal/event"
)

type EventCache struct {
	client redis.Cmdable
}

func NewEventCache(client redis.Cmdable) *EventCache {
	return &EventCache{client: client}
}

func eventListKey() string {
	return "cache:events:list:v1:ALL:*"
}

func eventDetailKey(eventID int64) string {
	return fmt.Sprintf("cache:event:detail:v1:%d", eventID)
}

func (c *EventCache) GetEventCard(ctx context.Context, eventID int64) (event.Card, bool, error) {
	payload, err := c.client.Get(ctx, eventDetailKey(eventID)).Result()
	if err == redis.Nil {
		return event.Card{}, false, nil
	}
	if err != nil {
		return event.Card{}, false, err
	}

	var card event.Card
	if err := json.Unmarshal([]byte(payload), &card); err != nil {
		return event.Card{}, false, err
	}
	return card, true, nil
}

func (c *EventCache) SetEventCard(ctx context.Context, eventID int64, card event.Card, status string) error {
	payload, err := json.Marshal(card)
	if err != nil {
		return err
	}
	ttl := 120 * time.Second
	if status == "live" {
		ttl = 20 * time.Second
	}
	return c.client.Set(ctx, eventDetailKey(eventID), payload, ttl).Err()
}

func (c *EventCache) GetEvents(ctx context.Context) ([]event.EventSummary, bool, error) {
	payload, err := c.client.Get(ctx, eventListKey()).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var items []event.EventSummary
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, false, err
	}
	return items, true, nil
}

func (c *EventCache) SetEvents(ctx context.Context, events []event.EventSummary) error {
	payload, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, eventListKey(), payload, 60*time.Second).Err()
}

func (c *EventCache) InvalidateEvent(ctx context.Context, eventID int64) error {
	return c.client.Del(ctx, eventDetailKey(eventID)).Err()
}

func (c *EventCache) InvalidateEvents(ctx context.Context) error {
	keys, err := c.client.Keys(ctx, "cache:events:list:v1:*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}
