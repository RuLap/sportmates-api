package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/RuLap/sportmates-api/internal/pkg/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
	log    *slog.Logger
}

func NewClient(cfg config.RedisConfig, log *slog.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("redis connected successfully", "address", cfg.Address)

	return &Client{
		client: rdb,
		log:    log,
	}, nil
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
