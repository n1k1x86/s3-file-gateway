package redis

import (
	"context"
	"fmt"
	"time"

	redis_v9 "github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string) (*redis_v9.Client, error) {
	client := redis_v9.NewClient(&redis_v9.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     20,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
