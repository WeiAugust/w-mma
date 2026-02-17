package queue

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestConsume_AcksMessageOnSuccess(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer client.Close()

	q := NewStreamQueue(client, "stream:test", "worker", "c1")
	ctx := context.Background()
	if err := q.EnsureGroup(ctx); err != nil {
		t.Fatalf("ensure group failed: %v", err)
	}
	if err := q.Publish(ctx, map[string]any{"source_id": "1", "url": "https://example.com/a"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	handled := false
	err = q.Consume(ctx, func(_ StreamMessage) error {
		handled = true
		return nil
	})
	if err != nil {
		t.Fatalf("consume failed: %v", err)
	}
	if !handled {
		t.Fatalf("message was not handled")
	}
}
