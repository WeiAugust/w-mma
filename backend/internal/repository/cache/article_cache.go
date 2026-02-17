package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const ArticlesListKey = "cache:articles:list:v1"

type ArticleCache struct {
	client redis.Cmdable
	ttl    time.Duration
}

func NewArticleCache(client redis.Cmdable, ttl time.Duration) *ArticleCache {
	if ttl <= 0 {
		ttl = 120 * time.Second
	}
	return &ArticleCache{client: client, ttl: ttl}
}

func (c *ArticleCache) InvalidateArticlesList(ctx context.Context) error {
	return c.client.Del(ctx, ArticlesListKey).Err()
}
