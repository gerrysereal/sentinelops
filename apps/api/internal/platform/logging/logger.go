package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	ActorKey     ctxKey = "actor"
)

func New(format string, output io.Writer) *slog.Logger {
	if output == nil {
		output = os.Stdout
	}
	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(handler)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	if value, ok := ctx.Value(RequestIDKey).(string); ok {
		return value
	}
	return ""
}

func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, ActorKey, actor)
}

func Actor(ctx context.Context) string {
	if value, ok := ctx.Value(ActorKey).(string); ok {
		return value
	}
	return "system"
}

func TraceID(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return ""
	}
	return spanContext.TraceID().String()
}

func SpanID(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return ""
	}
	return spanContext.SpanID().String()
}

// Attrs returns the standard correlation attributes for context-aware logs.
// Trace and span identifiers are added only when a valid OpenTelemetry span is
// present, so local no-op telemetry remains clean.
func Attrs(ctx context.Context) []slog.Attr {
	attrs := make([]slog.Attr, 0, 4)
	if requestID := RequestID(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	if actor := Actor(ctx); actor != "" {
		attrs = append(attrs, slog.String("actor", actor))
	}
	if traceID := TraceID(ctx); traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}
	if spanID := SpanID(ctx); spanID != "" {
		attrs = append(attrs, slog.String("span_id", spanID))
	}
	return attrs
}

func Since(start time.Time) slog.Attr {
	return slog.Int64("duration_ms", time.Since(start).Milliseconds())
}
