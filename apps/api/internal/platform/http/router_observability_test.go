package http

import (
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/auth"
)

func TestNewRouterAppliesObservabilityMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	observabilityMiddleware := func(c *gin.Context) {
		called = true
		c.Next()
	}

	cfg := config.Config{
		AppEnv:         config.EnvLocal,
		RequestTimeout: time.Second,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(
		cfg,
		nil,
		nil,
		nil,
		auth.NewMiddleware(cfg),
		nil,
		logger,
		observabilityMiddleware,
	)

	request := httptest.NewRequest(nethttp.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, response.Code)
	}
	if !called {
		t.Fatal("expected observability middleware to be called")
	}
}
