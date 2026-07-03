package cache

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/swarnava/dmb/services/search/internal/model"
)

type SearchCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSearchCache(redisAddr string, ttl time.Duration) *SearchCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &SearchCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *SearchCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *SearchCache) Close() error {
	return c.client.Close()
}

func (c *SearchCache) Get(ctx context.Context, filters model.SearchFilters) ([]model.SearchListing, bool, error) {
	key := buildSearchKey(filters)

	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get search cache: %w", err)
	}

	var results []model.SearchListing
	if err := json.Unmarshal([]byte(value), &results); err != nil {
		return nil, false, fmt.Errorf("failed to decode search cache: %w", err)
	}

	return results, true, nil
}

func (c *SearchCache) Set(ctx context.Context, filters model.SearchFilters, results []model.SearchListing) error {
	key := buildSearchKey(filters)

	data, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to encode search results: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set search cache: %w", err)
	}

	return nil
}

func buildSearchKey(filters model.SearchFilters) string {
	rawKey := fmt.Sprintf(
		"category=%d:min=%d:max=%d:limit=%d",
		filters.CategoryID,
		filters.MinPrice,
		filters.MaxPrice,
		filters.Limit,
	)

	hash := sha1.Sum([]byte(rawKey))

	return "search:" + hex.EncodeToString(hash[:])
}