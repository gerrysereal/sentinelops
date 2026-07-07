package config

import "os"

type Config struct {
	AppEnv            string
	APIPort           string
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	AuthEnabled       bool
	KeycloakIssuerURL string
	KeycloakClientID  string
}

func Load() Config {
	return Config{
		AppEnv:            getEnv("APP_ENV", "local"),
		APIPort:           getEnv("API_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://sentinelops:sentinelops@localhost:5432/sentinelops?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		AuthEnabled:       getEnv("AUTH_ENABLED", "false") == "true",
		KeycloakIssuerURL: getEnv("KEYCLOAK_ISSUER_URL", "http://localhost:8081/realms/sentinelops"),
		KeycloakClientID:  getEnv("KEYCLOAK_CLIENT_ID", "sentinelops-web"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
