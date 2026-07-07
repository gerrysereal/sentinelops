package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvLocal      = "local"
	EnvProduction = "production"

	ModeDemo       = "demo"
	ModeLab        = "lab"
	ModeProduction = "production"
)

var defaultSensitiveValues = []string{
	"sentinelops",
	"sentinelops-local-dev-key-change-me",
	"change-me",
	"changeme",
}

type Config struct {
	AppEnv             string
	APIPort            string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	AuthEnabled        bool
	BootstrapToken     string
	KeycloakIssuerURL  string
	KeycloakClientID   string
	PlatformMode       string
	EncryptionKey      string
	AllowedOrigins     []string
	RequestTimeout     time.Duration
	IntegrationTimeout time.Duration
	IntegrationRetries int
	LogFormat          string
}

func Load() Config {
	cfg := Config{
		AppEnv:             getEnv("APP_ENV", EnvLocal),
		APIPort:            getEnv("API_PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisAddr:          os.Getenv("REDIS_ADDR"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		AuthEnabled:        getEnvBool("AUTH_ENABLED", false),
		BootstrapToken:     os.Getenv("SENTINELOPS_BOOTSTRAP_TOKEN"),
		KeycloakIssuerURL:  os.Getenv("KEYCLOAK_ISSUER_URL"),
		KeycloakClientID:   os.Getenv("KEYCLOAK_CLIENT_ID"),
		PlatformMode:       getEnv("PLATFORM_MODE", ModeDemo),
		EncryptionKey:      os.Getenv("INTEGRATION_ENCRYPTION_KEY"),
		AllowedOrigins:     splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		RequestTimeout:     getEnvDuration("REQUEST_TIMEOUT", 15*time.Second),
		IntegrationTimeout: getEnvDuration("INTEGRATION_TIMEOUT", 8*time.Second),
		IntegrationRetries: getEnvInt("INTEGRATION_RETRIES", 2),
		LogFormat:          getEnv("LOG_FORMAT", "text"),
	}
	return cfg
}

func (c Config) Validate() error {
	var issues []string
	if c.APIPort == "" {
		issues = append(issues, "API_PORT is required")
	}
	if c.DatabaseURL == "" {
		issues = append(issues, "DATABASE_URL is required")
	} else if _, err := url.Parse(c.DatabaseURL); err != nil {
		issues = append(issues, fmt.Sprintf("DATABASE_URL is invalid: %v", err))
	}
	if c.RedisAddr == "" {
		issues = append(issues, "REDIS_ADDR is required")
	}
	if c.EncryptionKey == "" {
		issues = append(issues, "INTEGRATION_ENCRYPTION_KEY is required")
	}
	if !isAllowed(c.PlatformMode, ModeDemo, ModeLab, ModeProduction) {
		issues = append(issues, "PLATFORM_MODE must be one of demo, lab, production")
	}
	if !isAllowed(c.AppEnv, EnvLocal, "development", "test", "staging", EnvProduction) {
		issues = append(issues, "APP_ENV is invalid")
	}
	if c.IntegrationRetries < 0 || c.IntegrationRetries > 5 {
		issues = append(issues, "INTEGRATION_RETRIES must be between 0 and 5")
	}
	if c.RequestTimeout < time.Second {
		issues = append(issues, "REQUEST_TIMEOUT must be at least 1s")
	}
	if c.IntegrationTimeout < time.Second {
		issues = append(issues, "INTEGRATION_TIMEOUT must be at least 1s")
	}
	if c.IsProduction() {
		if !c.AuthEnabled {
			issues = append(issues, "AUTH_ENABLED must be true in production")
		}
		if isDefaultSensitive(c.EncryptionKey) {
			issues = append(issues, "INTEGRATION_ENCRYPTION_KEY must not use a default value in production")
		}
		if c.BootstrapToken == "" && strings.TrimSpace(c.KeycloakIssuerURL) == "" {
			issues = append(issues, "production auth requires SENTINELOPS_BOOTSTRAP_TOKEN or KEYCLOAK_ISSUER_URL")
		}
		if len(c.AllowedOrigins) == 0 {
			issues = append(issues, "CORS_ALLOWED_ORIGINS is required in production")
		}
		for _, origin := range c.AllowedOrigins {
			if origin == "*" {
				issues = append(issues, "wildcard CORS origin is not allowed in production")
			}
		}
	}
	if len(issues) > 0 {
		return errors.New(strings.Join(issues, "; "))
	}
	return nil
}

func (c Config) IsProduction() bool {
	return c.AppEnv == EnvProduction || c.PlatformMode == ModeProduction
}

func (c Config) PublicRuntimeConfig() map[string]any {
	return map[string]any{
		"appEnv":               c.AppEnv,
		"platformMode":         c.PlatformMode,
		"authEnabled":          c.AuthEnabled,
		"requestTimeoutMs":     c.RequestTimeout.Milliseconds(),
		"integrationTimeoutMs": c.IntegrationTimeout.Milliseconds(),
		"integrationRetries":   c.IntegrationRetries,
		"logFormat":            c.LogFormat,
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func isAllowed(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

func isDefaultSensitive(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range defaultSensitiveValues {
		if value == candidate {
			return true
		}
	}
	return false
}
