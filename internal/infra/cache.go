package infra

import (
	"context"
	"time"
)

type Cache interface {
	SetOne(ctx context.Context, key, value string, ttl time.Duration) error
	GetOne(ctx context.Context, key string) (string, error)
	GetAllKeysUsingWildCard(ctx context.Context, wildcard string) ([]string, error)
	DeleteOne(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
