// Package observability implements domain models for clickhouse
package observability

import (
	"context"
	"time"
)

type PerformanceLog struct {
	Endpoint   string
	Duration   time.Duration
	StatusCode int
	CacheHit   bool
}

type UserCH interface {
	SavePerformanceLog(ctx context.Context, log *PerformanceLog)
}
