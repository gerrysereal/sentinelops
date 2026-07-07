package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/sentinelops/sentinelops/apps/api/internal/application"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/auth"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/cache"
	secretcrypto "github.com/sentinelops/sentinelops/apps/api/internal/platform/crypto"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
	httpplatform "github.com/sentinelops/sentinelops/apps/api/internal/platform/http"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/integrations"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate postgres: %v", err)
	}

	if err := database.Seed(ctx, pool); err != nil {
		log.Fatalf("seed postgres: %v", err)
	}

	redisClient := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	defer redisClient.Close()

	repository := database.NewRepository(pool)
	overviewService := application.NewOverviewService(repository, redisClient)
	moduleService := application.NewModuleService(repository)

	secretBox, err := secretcrypto.NewSecretBox(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("initialize secret box: %v", err)
	}
	integrationRegistry := integrations.NewRegistry(cfg.PlatformMode)
	integrationService := application.NewIntegrationService(repository, secretBox, integrationRegistry, cfg.PlatformMode)

	authMiddleware := auth.NewMiddleware(cfg)
	router := httpplatform.NewRouter(cfg, overviewService, moduleService, integrationService, authMiddleware)

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("SentinelOps API listening on :%s", cfg.APIPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server failed: %v", err)
	}
}
