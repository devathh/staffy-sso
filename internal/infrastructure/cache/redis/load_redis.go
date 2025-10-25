package redis

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func ConnectToRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Secrets.Redis.Addr,
		DB:       cfg.Secrets.Redis.DB,
		Password: cfg.Secrets.Redis.Password,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return client, nil
}

func MustLoadCache(cfg *config.Config) *redis.Client {
	client, err := ConnectToRedis(cfg)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return client
}

func Close(client *redis.Client) error {
	return client.Conn().Close()
}
