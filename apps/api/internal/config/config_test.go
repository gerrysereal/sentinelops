package config

import "testing"

func TestValidateRequiresDatabaseURL(t *testing.T) {
	cfg := Config{
		AppEnv:             EnvLocal,
		APIPort:            "8080",
		RedisAddr:          "redis:6379",
		PlatformMode:       ModeDemo,
		EncryptionKey:      "test-encryption-key",
		RequestTimeout:     1,
		IntegrationTimeout: 1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing database URL to fail validation")
	}
}

func TestValidateRejectsProductionDefaults(t *testing.T) {
	cfg := Config{
		AppEnv:             EnvProduction,
		APIPort:            "8080",
		DatabaseURL:        "postgres://user:pass@postgres:5432/db",
		RedisAddr:          "redis:6379",
		PlatformMode:       ModeProduction,
		EncryptionKey:      "change-me",
		RequestTimeout:     1,
		IntegrationTimeout: 1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected production default secret validation failure")
	}
}
