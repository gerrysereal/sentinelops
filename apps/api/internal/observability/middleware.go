package observability

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// HTTPMiddleware returns the OpenTelemetry Gin middleware configured with the
// exact providers and propagator owned by this SDK. It instruments all routes
// registered after the middleware is attached and preserves incoming W3C
// trace context for distributed tracing.
func (s *SDK) HTTPMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(
		s.serviceName,
		otelgin.WithTracerProvider(s.tracerProvider),
		otelgin.WithMeterProvider(s.meterProvider),
		otelgin.WithPropagators(s.propagator),
	)
}
