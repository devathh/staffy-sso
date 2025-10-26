package clickhouse

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migrator struct {
	log  *slog.Logger
	conn driver.Conn
}

func (m *Migrator) Migrate(ctx context.Context) error {
	m.log.Info("starting clickhouse migrations")

	migrations, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	for _, migration := range migrations {
		if filepath.Ext(migration.Name()) != ".sql" {
			continue
		}

		m.log.Info("applying migration", slog.String("migration", migration.Name()))

		content, err := migrationFS.ReadFile("migrations/" + migration.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		if err := m.conn.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to exec migration: %w", err)
		}
	}

	m.log.Info("clickhouse migrations completed")

	return nil
}

func NewMigrator(log *slog.Logger, conn driver.Conn) *Migrator {
	return &Migrator{
		log:  log,
		conn: conn,
	}
}
