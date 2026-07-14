package observability

import (
	"errors"
	"fmt"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// InstrumentRedis attaches OpenTelemetry tracing and metrics hooks to a
// go-redis client. Raw Redis command statements are disabled so cache keys,
// values, tokens, and other command arguments are never exported.
func (s *SDK) InstrumentRedis(client redis.UniversalClient) error {
	if client == nil {
		return errors.New("redis client is required")
	}

	if err := redisotel.InstrumentTracing(
		client,
		redisotel.WithTracerProvider(s.tracerProvider),
		redisotel.WithDBStatement(false),
	); err != nil {
		return fmt.Errorf("instrument redis tracing: %w", err)
	}

	if err := redisotel.InstrumentMetrics(
		client,
		redisotel.WithMeterProvider(s.meterProvider),
	); err != nil {
		return fmt.Errorf("instrument redis metrics: %w", err)
	}

	return nil
}
