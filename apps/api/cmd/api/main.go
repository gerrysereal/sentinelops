package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sentinelops/sentinelops/apps/api/internal/application"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
	"github.com/sentinelops/sentinelops/apps/api/internal/observability"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/auth"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/cache"
	secretcrypto "github.com/sentinelops/sentinelops/apps/api/internal/platform/crypto"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
	httpplatform "github.com/sentinelops/sentinelops/apps/api/internal/platform/http"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/integrations"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/logging"
)

const (
	startupTimeout  = 30 * time.Second
	shutdownTimeout = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("SentinelOps API terminated", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	logger := logging.New(cfg.LogFormat, os.Stdout)
	slog.SetDefault(logger)
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid runtime configuration: %w", err)
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), startupTimeout)
	defer startupCancel()

	telemetryConfig := observability.LoadConfigFromEnv()
	telemetry, err := observability.New(startupCtx, telemetryConfig)
	if err != nil {
		return fmt.Errorf("initialize observability: %w", err)
	}
	defer shutdownTelemetry(logger, telemetry)

	logger.Info("observability initialized",
		"enabled", telemetryConfig.Enabled,
		"service_name", telemetryConfig.ServiceName,
		"service_version", telemetryConfig.ServiceVersion,
		"deployment_environment", telemetryConfig.DeploymentEnvironment,
		"otlp_endpoint", telemetryConfig.OTLPEndpoint,
	)

	pool, err := database.Connect(startupCtx, cfg.DatabaseURL, telemetry.PostgresTracer())
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	if err := telemetry.RecordPostgresStats(pool); err != nil {
		return fmt.Errorf("initialize postgres observability: %w", err)
	}

	if err := database.Migrate(startupCtx, pool); err != nil {
		return fmt.Errorf("migrate postgres: %w", err)
	}

	if err := database.Seed(startupCtx, pool, cfg.PlatformMode); err != nil {
		return fmt.Errorf("seed postgres: %w", err)
	}

	redisClient := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	defer redisClient.Close()

	repository := database.NewRepository(pool)
	overviewService := application.NewOverviewService(repository, redisClient)
	moduleService := application.NewModuleService(repository)

	secretBox, err := secretcrypto.NewSecretBox(cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("initialize secret box: %w", err)
	}
	integrationRegistry := integrations.NewRegistry(integrations.RegistryOptions{
		Mode:    cfg.PlatformMode,
		Timeout: cfg.IntegrationTimeout,
		Retries: cfg.IntegrationRetries,
	})
	integrationService := application.NewIntegrationService(repository, secretBox, integrationRegistry, cfg.PlatformMode)

	authMiddleware := auth.NewMiddleware(cfg)
	router := httpplatform.NewRouter(
		cfg,
		overviewService,
		moduleService,
		integrationService,
		authMiddleware,
		redisClient,
		logger,
		telemetry.HTTPMiddleware(),
	)

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("SentinelOps API listening", "port", cfg.APIPort, "app_env", cfg.AppEnv, "platform_mode", cfg.PlatformMode)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdownSignals)

	select {
	case signalValue := <-shutdownSignals:
		logger.Info("shutdown signal received", "signal", signalValue.String())
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("api server failed: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("api graceful shutdown: %w", err)
	}
	if err := telemetry.ForceFlush(shutdownCtx); err != nil {
		logger.Warn("observability force flush completed with errors", "error", err)
	}
	if err := telemetry.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown observability: %w", err)
	}

	logger.Info("SentinelOps API stopped")
	return nil
}

func shutdownTelemetry(logger *slog.Logger, telemetry *observability.SDK) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := telemetry.Shutdown(ctx); err != nil {
		logger.Error("observability shutdown failed", "error", err)
	}
}
