package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/olad5/go-hackathon-starter-template/config"
)

type RedisCache struct {
	Client  *redis.Client
	AppName *string
}

var ttl = time.Minute * 30

func New(ctx context.Context, configurations *config.Configurations) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: configurations.CacheAddress,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		Client:  client,
		AppName: &configurations.AppName,
	}, nil
}

func (r *RedisCache) SetOne(ctx context.Context, key, value string) error {
	_, err := r.Client.Set(ctx, *r.AppName+key, value, ttl).Result()
	if err != nil {
		return fmt.Errorf("Error setting value in cache: %w", err)
	}
	return nil
}

func (r *RedisCache) GetOne(ctx context.Context, key string) (string, error) {
	result, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("Error getting value from cache: %w", err)
	}
	return result, nil
}

func (r *RedisCache) DeleteOne(ctx context.Context, key string) error {
	_, err := r.Client.Del(ctx, key).Result()
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
