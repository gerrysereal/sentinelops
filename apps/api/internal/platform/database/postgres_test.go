package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

type testQueryTracer struct{}

func (testQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (testQueryTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestPoolConfigAttachesQueryTracer(t *testing.T) {
	tracer := testQueryTracer{}
	cfg, err := poolConfig("postgres://user:password@localhost:5432/sentinelops?sslmode=disable", tracer)
	if err != nil {
		t.Fatalf("create pool config: %v", err)
	}
	if _, ok := cfg.ConnConfig.Tracer.(testQueryTracer); !ok {
		t.Fatalf("expected test query tracer, got %T", cfg.ConnConfig.Tracer)
	}
}
