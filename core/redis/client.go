package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string
	Password string
	DB       int
}

func Open(ctx context.Context, options Options) (*goredis.Client, error) {
	if options.Addr == "" {
		return nil, nil
	}
	client := goredis.NewClient(&goredis.Options{Addr: options.Addr, Password: options.Password, DB: options.DB})
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
