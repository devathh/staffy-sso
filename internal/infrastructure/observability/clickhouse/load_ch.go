// Package clickhouse implements connecting to ch n' interacting with it
package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/devathh/staffy-sso/internal/infrastructure/config"
)

func ConnectToCH(ctx context.Context, cfg *config.Config) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Secrets.Clickhouse.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Secrets.Clickhouse.Database,
			Password: cfg.Secrets.Clickhouse.Password,
			Username: cfg.Secrets.Clickhouse.Username,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return conn, nil
}

func MustConnectToCH(ctx context.Context, cfg *config.Config) driver.Conn {
	conn, err := ConnectToCH(ctx, cfg)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return conn
}

func Close(conn driver.Conn) error {
	return conn.Close()
}
