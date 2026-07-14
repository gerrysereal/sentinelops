package logging

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestAttrsIncludesRequestActorAndTraceCorrelation(t *testing.T) {
	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
	ctx = WithRequestID(ctx, "request-123")
	ctx = WithActor(ctx, "platform-admin")

	values := map[string]string{}
	for _, attr := range Attrs(ctx) {
		values[attr.Key] = attr.Value.String()
	}

	expected := map[string]string{
		"request_id": "request-123",
		"actor":      "platform-admin",
		"trace_id":   traceID.String(),
		"span_id":    spanID.String(),
	}
	for key, value := range expected {
		if values[key] != value {
			t.Fatalf("unexpected %s: got %q want %q", key, values[key], value)
		}
	}
}

func TestAttrsOmitsTraceCorrelationWithoutValidSpan(t *testing.T) {
	ctx := WithRequestID(context.Background(), "request-123")
	for _, attr := range Attrs(ctx) {
		if attr.Key == "trace_id" || attr.Key == "span_id" {
			t.Fatalf("unexpected %s without valid span", attr.Key)
		}
	}
}
