package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/olad5/caution-companion/config"
)

type RedisCache struct {
	Client  *redis.Client
	AppName *string
}

func New(ctx context.Context, configurations *config.Configurations) (*RedisCache, error) {
	opts, err := redis.ParseURL(configurations.CacheAddress)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse redis url: %v", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		Client:  client,
		AppName: &configurations.AppName,
	}, nil
}

func (r *RedisCache) SetOne(ctx context.Context, key, value string, ttl time.Duration) error {
	_, err := r.Client.Set(ctx, r.prefixKeyWithAppName(key), value, ttl).Result()
	if err != nil {
		return fmt.Errorf("Error setting value in cache: %w", err)
	}
	return nil
}

func (r *RedisCache) GetOne(ctx context.Context, key string) (string, error) {
	result, err := r.Client.Get(ctx, r.prefixKeyWithAppName(key)).Result()
	if err != nil {
		return "", fmt.Errorf("Error getting value from cache: %w", err)
	}
	return result, nil
}

func (r *RedisCache) GetAllKeysUsingWildCard(ctx context.Context, wildcard string) ([]string, error) {
	rr, err := r.Client.Keys(ctx, wildcard).Result()
	if err != nil {
		return []string{""}, fmt.Errorf("Error getting wildcard values from cache: %w", err)
	}
	var results []string
	for _, result := range rr {
		results = append(results, r.removeAppNamePrefixKey(result))
	}
	return results, nil
}

func (r *RedisCache) DeleteOne(ctx context.Context, key string) error {
	_, err := r.Client.Del(ctx, r.prefixKeyWithAppName(key)).Result()
	if err != nil {
		return fmt.Errorf("Error deleting key in cache: %w", err)
	}
	return nil
}

func (r *RedisCache) Ping(ctx context.Context) error {
	if err := r.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Failed to Ping Redis Cache: %v", err)
	}
	return nil
}

func (r *RedisCache) removeAppNamePrefixKey(key string) string {
	return strings.TrimPrefix(key, *r.AppName)
}

func (r *RedisCache) prefixKeyWithAppName(key string) string {
	return *r.AppName + key
}
