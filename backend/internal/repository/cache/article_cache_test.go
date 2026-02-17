package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestInvalidateArticlesList(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer client.Close()

	cache := NewArticleCache(client, time.Minute)
	ctx := context.Background()

	if err := client.Set(ctx, ArticlesListKey, `{"items":[1]}`, time.Minute).Err(); err != nil {
		t.Fatalf("seed cache failed: %v", err)
	}

	if err := cache.InvalidateArticlesList(ctx); err != nil {
		t.Fatalf("invalidate cache failed: %v", err)
	}

	if mini.Exists(ArticlesListKey) {
		t.Fatalf("expected key %q to be deleted", ArticlesListKey)
	}
}
