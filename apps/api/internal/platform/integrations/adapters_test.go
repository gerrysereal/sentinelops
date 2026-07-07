package integrations

import (
	"context"
	"testing"
	"time"

	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
)

func TestHTTPAdapterDisabledIntegrationDoesNotFakeSuccess(t *testing.T) {
	adapter := HTTPAdapter{client: httpClient(true, time.Second), retries: 0}
	health, err := adapter.Health(context.Background(), domain.IntegrationConfig{Enabled: false, Type: "Prometheus"})
	if err != nil {
		t.Fatalf("health returned error: %v", err)
	}
	if health.Healthy || health.Status != domain.IntegrationStatusDisabled {
		t.Fatalf("expected disabled health, got healthy=%v status=%s", health.Healthy, health.Status)
	}
}
