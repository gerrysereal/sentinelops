package observability

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const postgresStatsInterval = 15 * time.Second

// PostgresTracer returns a pgx query tracer that uses the same tracer and
// meter providers as SentinelOps HTTP instrumentation. SQL text, query
// parameters, and connection details are deliberately excluded from spans.
func (s *SDK) PostgresTracer() pgx.QueryTracer {
	return otelpgx.NewTracer(
		otelpgx.WithTracerProvider(s.tracerProvider),
		otelpgx.WithMeterProvider(s.meterProvider),
		otelpgx.WithSpanNameFunc(postgresSpanName),
		otelpgx.WithTrimSQLInSpanName(),
		otelpgx.WithDisableQuerySpanNamePrefix(),
		otelpgx.WithDisableConnectionDetailsInAttributes(),
		otelpgx.WithDisableSQLStatementInAttributes(),
	)
}

// RecordPostgresStats registers connection-pool metrics with the SDK meter
// provider. The callback has the same process lifetime as the pool.
func (s *SDK) RecordPostgresStats(pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("postgres pool is required")
	}
	if err := otelpgx.RecordStats(
		pool,
		otelpgx.WithStatsMeterProvider(s.meterProvider),
		otelpgx.WithMinimumReadDBStatsInterval(postgresStatsInterval),
	); err != nil {
		return fmt.Errorf("register postgres pool metrics: %w", err)
	}
	return nil
}

func postgresSpanName(statement string) string {
	fields := strings.Fields(statement)
	if len(fields) == 0 {
		return "postgresql query"
	}

	operation := strings.ToUpper(fields[0])
	switch operation {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "UPSERT", "MERGE", "CALL", "COPY", "BEGIN", "COMMIT", "ROLLBACK":
		return "postgresql " + operation
	default:
		return "postgresql query"
	}
}
