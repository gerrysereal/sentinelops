package observability

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestPostgresSpanNameExposesOnlyOperation(t *testing.T) {
	statement := `SELECT access_token, password FROM integrations WHERE id = $1`
	name := postgresSpanName(statement)

	if name != "postgresql SELECT" {
		t.Fatalf("unexpected span name: %q", name)
	}
	for _, sensitive := range []string{"access_token", "password", "integrations", "$1"} {
		if strings.Contains(strings.ToLower(name), strings.ToLower(sensitive)) {
			t.Fatalf("span name leaked %q: %q", sensitive, name)
		}
	}
}

func TestPostgresSpanNameFallsBackForUnknownStatement(t *testing.T) {
	if got := postgresSpanName("/* comment */ SELECT 1"); got != "postgresql query" {
		t.Fatalf("unexpected fallback span name: %q", got)
	}
}

func TestDisabledSDKProvidesNoopCompatiblePostgresTracer(t *testing.T) {
	sdk, err := New(context.Background(), Config{
		Enabled:              false,
		ServiceName:          DefaultServiceName,
		ExportTimeout:        time.Second,
		BatchTimeout:         time.Second,
		MetricExportInterval: time.Second,
		Sampler:              "parentbased_always_on",
		SamplerArgument:      1,
	})
	if err != nil {
		t.Fatalf("initialize disabled SDK: %v", err)
	}
	if sdk.PostgresTracer() == nil {
		t.Fatal("expected postgres tracer")
	}
	if err := sdk.RecordPostgresStats(nil); err == nil {
		t.Fatal("expected nil pool validation error")
	}
}

func TestPostgresTracerDoesNotExposeSQLInSpanName(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(recorder)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	sdk := &SDK{
		tracerProvider: provider,
		meterProvider:  newNoopMeterProvider(),
	}
	tracer := sdk.PostgresTracer()

	ctx, root := provider.Tracer("sentinelops-test").Start(context.Background(), "request")
	ctx = tracer.TraceQueryStart(ctx, nil, pgx.TraceQueryStartData{
		SQL:  `SELECT access_token, password FROM integrations WHERE id = $1`,
		Args: []any{"secret-integration-id"},
	})
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		CommandTag: pgconn.NewCommandTag("SELECT 1"),
	})
	root.End()

	var querySpanName string
	for _, span := range recorder.Ended() {
		if span.SpanKind().String() == "client" {
			querySpanName = span.Name()
			break
		}
	}
	if querySpanName != "postgresql SELECT" {
		t.Fatalf("unexpected PostgreSQL span name: %q", querySpanName)
	}

	for _, sensitive := range []string{
		"access_token",
		"password",
		"integrations",
		"$1",
		"secret-integration-id",
	} {
		if strings.Contains(strings.ToLower(querySpanName), strings.ToLower(sensitive)) {
			t.Fatalf("span name leaked %q: %q", sensitive, querySpanName)
		}
	}
}
