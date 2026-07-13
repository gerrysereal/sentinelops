package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPMiddlewareCreatesServerSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		_ = tracerProvider.Shutdown(context.Background())
	})

	sdk := &SDK{
		serviceName:    DefaultServiceName,
		tracerProvider: tracerProvider,
		meterProvider:  noop.NewMeterProvider(),
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}

	router := gin.New()
	router.Use(sdk.HTTPMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("unexpected response status: %d", response.Code)
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected one HTTP server span, got %d", len(spans))
	}
	if spans[0].Name == "" {
		t.Fatal("expected HTTP server span to have a name")
	}
}
