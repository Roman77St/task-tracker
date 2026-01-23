package repository

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func NewRedisRepo(addr string) *RedisRepo {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = addr
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisRepo{Client: rdb}
}

func (r *RedisRepo) SetToken(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisRepo) GetToken(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisRepo) DeleteToken(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
