package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
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

func Attrs(ctx context.Context) []slog.Attr {
	attrs := []slog.Attr{}
	if requestID := RequestID(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	if actor := Actor(ctx); actor != "" {
		attrs = append(attrs, slog.String("actor", actor))
	}
	return attrs
}

func Since(start time.Time) slog.Attr {
	return slog.Int64("duration_ms", time.Since(start).Milliseconds())
}
