package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	platformlogging "github.com/sentinelops/sentinelops/apps/api/internal/platform/logging"
	"go.opentelemetry.io/otel/trace"
)

func TestRequestLoggerIncludesTraceAndSpanIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	var output bytes.Buffer
	logger := platformlogging.New("json", &output)
	router := gin.New()
	router.Use(requestID())
	router.Use(func(c *gin.Context) {
		ctx := trace.ContextWithSpanContext(c.Request.Context(), spanContext)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(requestLogger(logger))
	router.Use(func(c *gin.Context) {
		ctx := platformlogging.WithActor(c.Request.Context(), "local-admin")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("X-Request-ID", "request-123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", response.Code)
	}

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("decode structured log: %v\n%s", err, output.String())
	}

	expected := map[string]string{
		"request_id": "request-123",
		"actor":      "local-admin",
		"trace_id":   traceID.String(),
		"span_id":    spanID.String(),
	}
	for key, value := range expected {
		if entry[key] != value {
			t.Fatalf("unexpected %s: got %v want %q", key, entry[key], value)
		}
	}
}
