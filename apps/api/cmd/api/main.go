package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sentinelops/sentinelops/apps/api/internal/application"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/auth"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/cache"
	secretcrypto "github.com/sentinelops/sentinelops/apps/api/internal/platform/crypto"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
	httpplatform "github.com/sentinelops/sentinelops/apps/api/internal/platform/http"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/integrations"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/logging"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogFormat, os.Stdout)
	slog.SetDefault(logger)
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid runtime configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		logger.Error("migrate postgres failed", "error", err)
		os.Exit(1)
	}

	if err := database.Seed(ctx, pool, cfg.PlatformMode); err != nil {
		logger.Error("seed postgres failed", "error", err)
		os.Exit(1)
	}

	redisClient := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	defer redisClient.Close()

	repository := database.NewRepository(pool)
	overviewService := application.NewOverviewService(repository, redisClient)
	moduleService := application.NewModuleService(repository)

	secretBox, err := secretcrypto.NewSecretBox(cfg.EncryptionKey)
	if err != nil {
		logger.Error("initialize secret box failed", "error", err)
		os.Exit(1)
	}
	integrationRegistry := integrations.NewRegistry(integrations.RegistryOptions{
		Mode:    cfg.PlatformMode,
		Timeout: cfg.IntegrationTimeout,
		Retries: cfg.IntegrationRetries,
	})
	integrationService := application.NewIntegrationService(repository, secretBox, integrationRegistry, cfg.PlatformMode)

	authMiddleware := auth.NewMiddleware(cfg)
	router := httpplatform.NewRouter(cfg, overviewService, moduleService, integrationService, authMiddleware, redisClient, logger)

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("SentinelOps API listening", "port", cfg.APIPort, "app_env", cfg.AppEnv, "platform_mode", cfg.PlatformMode)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("api graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("SentinelOps API stopped")
}
