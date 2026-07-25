package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	platformlogging "github.com/sentinelops/sentinelops/apps/api/internal/platform/logging"
)

func TestRespondWithStatusSanitizesInternalErrorAndLogsCause(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := platformlogging.New("json", &logs)

	router := gin.New()
	router.Use(requestID())
	router.Use(requestLogger(logger))

	internalErr := errors.New(
		"failed to connect user=sentinelops database=sentinelops host=postgres",
	)

	router.GET("/test", func(c *gin.Context) {
		respond(c, nil, internalErr)
	})

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("X-Request-ID", "request-error-test")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"unexpected status: got %d want %d",
			response.Code,
			http.StatusInternalServerError,
		)
	}

	var body struct {
		Error     string `json:"error"`
		RequestID string `json:"requestId"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Error != "internal server error" {
		t.Fatalf(
			"unexpected public error: %q",
			body.Error,
		)
	}

	if body.RequestID != "request-error-test" {
		t.Fatalf(
			"unexpected request ID: %q",
			body.RequestID,
		)
	}

	if strings.Contains(response.Body.String(), "sentinelops") ||
		strings.Contains(response.Body.String(), "postgres") {
		t.Fatalf(
			"response leaks internal details: %s",
			response.Body.String(),
		)
	}

	var entry map[string]any
	if err := json.Unmarshal(
		bytes.TrimSpace(logs.Bytes()),
		&entry,
	); err != nil {
		t.Fatalf(
			"decode structured log: %v\n%s",
			err,
			logs.String(),
		)
	}

	if entry["request_id"] != "request-error-test" {
		t.Fatalf(
			"log missing request ID: %v",
			entry["request_id"],
		)
	}

	if entry["error"] != internalErr.Error() {
		t.Fatalf(
			"log missing internal error: got %v want %q",
			entry["error"],
			internalErr.Error(),
		)
	}
}

func TestRespondWithStatusSanitizesTimeoutError(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Writer.Header().Set(
		"X-Request-ID",
		"request-timeout-test",
	)

	respond(c, nil, context.DeadlineExceeded)

	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf(
			"unexpected status: got %d want %d",
			response.Code,
			http.StatusGatewayTimeout,
		)
	}

	var body struct {
		Error     string `json:"error"`
		RequestID string `json:"requestId"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Error != "request timed out" {
		t.Fatalf(
			"unexpected public error: %q",
			body.Error,
		)
	}

	if strings.Contains(
		response.Body.String(),
		context.DeadlineExceeded.Error(),
	) {
		t.Fatalf(
			"response leaks internal timeout error: %s",
			response.Body.String(),
		)
	}
}
