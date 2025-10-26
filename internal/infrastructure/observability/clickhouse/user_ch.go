package clickhouse

import (
	"context"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/devathh/staffy-sso/internal/domain/observability"
	"github.com/devathh/staffy-sso/pkg/consts"
)

type UserCH struct {
	conn driver.Conn
	log  *slog.Logger
}

func (u *UserCH) SavePerformanceLog(ctx context.Context, log *observability.PerformanceLog) {
	defer func() {
		if r := recover(); r != nil {
			u.log.Warn("panic in performance log",
				slog.Any("panic", r),
				slog.String("endpoint", log.Endpoint))
		}
	}()

	if err := ctx.Err(); err != nil {
		u.log.Debug("context error", slog.String("error", err.Error()))
		return
	}

	err := u.conn.Exec(ctx, `INSERT INTO performance_logs (
			endpoint,
			duration,
			status_code,
			cache_hit
		) VALUES (?, ?, ?, ?)`, log.Endpoint, int64(log.Duration), log.StatusCode, log.CacheHit)

	if err != nil {
		u.log.Error("failed to insert log to performance_log", slog.String("error", err.Error()),
			slog.String("endpoint", log.Endpoint))
	}
}

func NewUserCH(log *slog.Logger, conn driver.Conn) (*UserCH, error) {
	if conn == nil || log == nil {
		return nil, consts.ErrInvalidArgs
	}

	return &UserCH{
		log:  log,
		conn: conn,
	}, nil
}
