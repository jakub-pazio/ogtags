package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	GetRequiredTags(ctx context.Context, url string) (bool, string, error)
	SetRequiredTags(ctx context.Context, url, body string) error
}

var _ Cache = (*RedisCache)(nil)

type RedisCache struct {
	client *redis.Client
}

func (r *RedisCache) GetRequiredTags(ctx context.Context, url string) (bool, string, error) {
	result, err := r.client.GetEx(ctx, url, time.Minute).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, "", nil
		} else {
			return false, "", err
		}
	}

	return true, result, nil
}

func (r *RedisCache) SetRequiredTags(ctx context.Context, url string, requiredTags string) error {
	return r.client.Set(ctx, url, requiredTags, time.Minute).Err()
}

func New() (Cache, func() error) {
	//TODO: move this config into arguments
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &RedisCache{rdb}, rdb.Close
}
