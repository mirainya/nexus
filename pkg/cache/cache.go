package cache

import (
	"context"
	"time"

	"github.com/mirainya/nexus/pkg/config"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func Init() {
	cfg := config.C.Redis
	if cfg.Addr == "" {
		return
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
}

func Available() bool {
	return rdb != nil
}

func Get(ctx context.Context, key string) (string, error) {
	if rdb == nil {
		return "", redis.Nil
	}
	return rdb.Get(ctx, key).Result()
}

func Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	return rdb.Set(ctx, key, value, ttl).Err()
}

func Del(ctx context.Context, key string) error {
	if rdb == nil {
		return nil
	}
	return rdb.Del(ctx, key).Err()
}
