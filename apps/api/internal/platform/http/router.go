package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sentinelops/sentinelops/apps/api/internal/application"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/auth"
)

type Handler struct {
	cfg          config.Config
	overview     *application.OverviewService
	modules      *application.ModuleService
	integrations *application.IntegrationService
	redis        *redis.Client
}

func NewRouter(cfg config.Config, overview *application.OverviewService, modules *application.ModuleService, integrations *application.IntegrationService, authMiddleware *auth.Middleware, redisClient *redis.Client, logger *slog.Logger) *gin.Engine {
	if cfg.AppEnv == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestID())
	router.Use(requestTimeout(cfg.RequestTimeout))
	router.Use(requestLogger(logger))
	if len(cfg.AllowedOrigins) > 0 {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.AllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-Role", "X-Request-ID"},
			ExposeHeaders:    []string{"X-Request-ID"},
			AllowCredentials: true,
		}))
	}

	h := &Handler{cfg: cfg, overview: overview, modules: modules, integrations: integrations, redis: redisClient}

	router.GET("/health", h.health)
	router.GET("/ready", h.readiness)
	router.GET("/live", h.liveness)

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth())
	{
		v1.GET("/platform/config", authMiddleware.RequirePermission("platform:read"), h.platformConfig)
		v1.GET("/platform/health", authMiddleware.RequirePermission("platform:read"), h.readiness)
		v1.GET("/platform/readiness", authMiddleware.RequirePermission("platform:read"), h.readiness)
		v1.GET("/platform/liveness", authMiddleware.RequirePermission("platform:read"), h.liveness)

		v1.GET("/overview", authMiddleware.RequirePermission("platform:read"), h.getOverview)
		v1.GET("/applications", authMiddleware.RequirePermission("application:read"), h.listApplications)
		v1.POST("/applications", authMiddleware.RequirePermission("application:write"), h.createApplication)
		v1.GET("/pipelines", authMiddleware.RequirePermission("application:read"), h.listPipelines)
		v1.POST("/pipelines/run", authMiddleware.RequirePermission("pipeline:operate"), h.runPipeline)
		v1.GET("/deployments", authMiddleware.RequirePermission("application:read"), h.listDeployments)
		v1.POST("/deployments", authMiddleware.RequirePermission("deployment:operate"), h.createDeployment)
		v1.GET("/security/alerts", authMiddleware.RequirePermission("application:read"), h.listSecurityAlerts)
		v1.POST("/security/scans", authMiddleware.RequirePermission("security:operate"), h.runSecurityScan)
		v1.PATCH("/security/alerts/:id/status", authMiddleware.RequirePermission("security:operate"), h.updateSecurityAlertStatus)
		v1.GET("/observability/signals", authMiddleware.RequirePermission("platform:read"), h.listObservabilitySignals)
		v1.GET("/registry/artifacts", authMiddleware.RequirePermission("platform:read"), h.listRegistryArtifacts)

		v1.GET("/settings", authMiddleware.RequirePermission("settings:read"), h.listSettings)
		v1.PUT("/settings/:key", authMiddleware.RequirePermission("settings:write"), h.upsertSetting)

		v1.GET("/integration-types", authMiddleware.RequirePermission("integration:read"), h.integrationTypes)
		v1.GET("/integrations", authMiddleware.RequirePermission("integration:read"), h.listIntegrations)
		v1.POST("/integrations", authMiddleware.RequirePermission("integration:write"), h.createIntegration)
		v1.GET("/integrations/:id", authMiddleware.RequirePermission("integration:read"), h.getIntegration)
		v1.PUT("/integrations/:id", authMiddleware.RequirePermission("integration:write"), h.updateIntegration)
		v1.DELETE("/integrations/:id", authMiddleware.RequirePermission("integration:delete"), h.deleteIntegration)
		v1.PATCH("/integrations/:id/enabled", authMiddleware.RequirePermission("integration:operate"), h.setIntegrationEnabled)
		v1.POST("/integrations/:id/test", authMiddleware.RequirePermission("integration:operate"), h.testIntegration)
		v1.POST("/integrations/:id/sync", authMiddleware.RequirePermission("integration:operate"), h.syncIntegration)
		v1.GET("/integrations/:id/status", authMiddleware.RequirePermission("integration:read"), h.getIntegrationStatus)
		v1.GET("/integrations/:id/logs", authMiddleware.RequirePermission("integration:read"), h.listIntegrationLogs)
		v1.GET("/integrations/:id/connection-history", authMiddleware.RequirePermission("integration:read"), h.listConnectionHistory)
	}

	return router
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sentinelops-api"})
}

func (h *Handler) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive", "checkedAt": time.Now().UTC()})
}

func (h *Handler) readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	checks := gin.H{"api": "ok"}
	ready := true
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = err.Error()
			ready = false
		} else {
			checks["redis"] = "ok"
		}
	}
	status := http.StatusOK
	state := "ready"
	if !ready {
		status = http.StatusServiceUnavailable
		state = "degraded"
	}
	c.JSON(status, gin.H{"status": state, "checks": checks, "checkedAt": time.Now().UTC()})
}

func (h *Handler) platformConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.cfg.PublicRuntimeConfig())
}

func (h *Handler) getOverview(c *gin.Context) {
	data, err := h.overview.GetOverview(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listApplications(c *gin.Context) {
	data, err := h.modules.ListApplications(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) createApplication(c *gin.Context) {
	var req domain.CreateApplicationRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.modules.CreateApplication(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) listPipelines(c *gin.Context) {
	data, err := h.modules.ListPipelines(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) runPipeline(c *gin.Context) {
	var req domain.CreatePipelineRunRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.modules.RunPipeline(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) listDeployments(c *gin.Context) {
	data, err := h.modules.ListDeployments(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) createDeployment(c *gin.Context) {
	var req domain.CreateDeploymentRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.modules.CreateDeployment(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) listSecurityAlerts(c *gin.Context) {
	data, err := h.modules.ListSecurityAlerts(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) runSecurityScan(c *gin.Context) {
	var req domain.CreateSecurityScanRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.modules.RunSecurityScan(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) updateSecurityAlertStatus(c *gin.Context) {
	var req domain.UpdateAlertStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.modules.UpdateSecurityAlertStatus(c.Request.Context(), c.Param("id"), req)
	h.overview.Invalidate(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listObservabilitySignals(c *gin.Context) {
	data, err := h.modules.ListObservabilitySignals(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listRegistryArtifacts(c *gin.Context) {
	data, err := h.modules.ListRegistryArtifacts(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listSettings(c *gin.Context) {
	data, err := h.modules.ListSettings(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) upsertSetting(c *gin.Context) {
	var req domain.UpdateSettingRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.modules.UpsertSetting(c.Request.Context(), c.Param("key"), req)
	respond(c, data, err)
}

func (h *Handler) integrationTypes(c *gin.Context) {
	respond(c, h.integrations.Types(c.Request.Context()), nil)
}

func (h *Handler) listIntegrations(c *gin.Context) {
	data, err := h.integrations.List(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) getIntegration(c *gin.Context) {
	data, err := h.integrations.Get(c.Request.Context(), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) createIntegration(c *gin.Context) {
	var req domain.CreateIntegrationRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.integrations.Create(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) updateIntegration(c *gin.Context) {
	var req domain.UpdateIntegrationRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.integrations.Update(c.Request.Context(), c.Param("id"), req)
	h.overview.Invalidate(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) deleteIntegration(c *gin.Context) {
	if err := h.integrations.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respond(c, nil, err)
		return
	}
	h.overview.Invalidate(c.Request.Context())
	c.Status(http.StatusNoContent)
}

func (h *Handler) setIntegrationEnabled(c *gin.Context) {
	var req domain.SetIntegrationEnabledRequest
	if !bindJSON(c, &req) {
		return
	}
	data, err := h.integrations.SetEnabled(c.Request.Context(), c.Param("id"), req)
	h.overview.Invalidate(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) testIntegration(c *gin.Context) {
	data, err := h.integrations.TestConnection(c.Request.Context(), c.Param("id"))
	h.overview.Invalidate(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) syncIntegration(c *gin.Context) {
	data, err := h.integrations.Sync(c.Request.Context(), c.Param("id"))
	h.overview.Invalidate(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) getIntegrationStatus(c *gin.Context) {
	data, err := h.integrations.Status(c.Request.Context(), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listIntegrationLogs(c *gin.Context) {
	data, err := h.integrations.Logs(c.Request.Context(), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listConnectionHistory(c *gin.Context) {
	data, err := h.integrations.ConnectionHistory(c.Request.Context(), c.Param("id"))
	respond(c, data, err)
}

func bindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "requestId": c.Writer.Header().Get("X-Request-ID")})
		return false
	}
	return true
}

func respond(c *gin.Context, data any, err error) {
	respondWithStatus(c, http.StatusOK, data, err)
}

func respondWithStatus(c *gin.Context, status int, data any, err error) {
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			statusCode = http.StatusGatewayTimeout
		}
		c.JSON(statusCode, gin.H{"error": err.Error(), "requestId": c.Writer.Header().Get("X-Request-ID")})
		return
	}
	c.JSON(status, data)
}
