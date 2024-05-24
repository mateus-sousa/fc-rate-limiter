package repository

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type LimiterConfig struct {
	FirstRequestTime      time.Time `json:"first_request_time"`
	AmountRequestInSecond int       `json:"amount_request_in_second"`
	Blocked               bool      `json:"blocked"`
	BlockedTime           time.Time `json:"blocked_time"`
}

type LimiterConfigRepository interface {
	GetRequestsBy(ctx context.Context, key string) (string, error)
	SetRequestsAmount(ctx context.Context, key string, limiterConfig []byte) error
}

type LimiterConfigCacheRepository struct {
	client *redis.Client
}

func NewLimiterConfigCacheRepository(client *redis.Client) *LimiterConfigCacheRepository {
	return &LimiterConfigCacheRepository{client: client}
}

func (l *LimiterConfigCacheRepository) GetRequestsBy(ctx context.Context, key string) (string, error) {
	val, err := l.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (l *LimiterConfigCacheRepository) SetRequestsAmount(ctx context.Context, key string, limiterConfig []byte) error {
	err := l.client.Set(ctx, key, limiterConfig, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
