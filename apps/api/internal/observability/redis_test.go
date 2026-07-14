package observability

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestInstrumentRedisRejectsNilClient(t *testing.T) {
	sdk := &SDK{}
	if err := sdk.InstrumentRedis(nil); err == nil {
		t.Fatal("expected nil Redis client validation error")
	}
}

func TestDisabledSDKCanInstrumentRedisClient(t *testing.T) {
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

	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() { _ = client.Close() })

	if err := sdk.InstrumentRedis(client); err != nil {
		t.Fatalf("instrument Redis client: %v", err)
	}
}
