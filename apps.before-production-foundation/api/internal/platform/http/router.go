package http

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sentinelops/sentinelops/apps/api/internal/application"
	"github.com/sentinelops/sentinelops/apps/api/internal/config"
	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/auth"
)

type Handler struct {
	overview     *application.OverviewService
	modules      *application.ModuleService
	integrations *application.IntegrationService
}

func NewRouter(cfg config.Config, overview *application.OverviewService, modules *application.ModuleService, integrations *application.IntegrationService, authMiddleware *auth.Middleware) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestID())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-Role"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))

	h := &Handler{overview: overview, modules: modules, integrations: integrations}

	router.GET("/health", h.health)

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth())
	{
		v1.GET("/overview", h.getOverview)
		v1.GET("/applications", h.listApplications)
		v1.POST("/applications", h.createApplication)
		v1.GET("/pipelines", h.listPipelines)
		v1.POST("/pipelines/run", h.runPipeline)
		v1.GET("/deployments", h.listDeployments)
		v1.POST("/deployments", h.createDeployment)
		v1.GET("/security/alerts", h.listSecurityAlerts)
		v1.POST("/security/scans", h.runSecurityScan)
		v1.PATCH("/security/alerts/:id/status", h.updateSecurityAlertStatus)
		v1.GET("/observability/signals", h.listObservabilitySignals)
		v1.GET("/integrations", h.listIntegrations)
		v1.POST("/integrations", h.createIntegration)
		v1.GET("/integrations/:id", h.getIntegration)
		v1.PUT("/integrations/:id", h.updateIntegration)
		v1.DELETE("/integrations/:id", h.deleteIntegration)
		v1.PATCH("/integrations/:id/enabled", h.setIntegrationEnabled)
		v1.POST("/integrations/:id/test", h.testIntegration)
		v1.POST("/integrations/:id/sync", h.syncIntegration)
		v1.GET("/integrations/:id/logs", h.listIntegrationLogs)
		v1.GET("/registry/artifacts", h.listRegistryArtifacts)
	}

	return router
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sentinelops-api"})
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.modules.RunSecurityScan(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) updateSecurityAlertStatus(c *gin.Context) {
	var req domain.UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.integrations.Create(c.Request.Context(), req)
	h.overview.Invalidate(c.Request.Context())
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) updateIntegration(c *gin.Context) {
	var req domain.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.integrations.Update(c.Request.Context(), c.Param("id"), req)
	h.overview.Invalidate(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) deleteIntegration(c *gin.Context) {
	if err := h.integrations.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.overview.Invalidate(c.Request.Context())
	c.Status(http.StatusNoContent)
}

func (h *Handler) setIntegrationEnabled(c *gin.Context) {
	var req domain.SetIntegrationEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func (h *Handler) listIntegrationLogs(c *gin.Context) {
	data, err := h.integrations.Logs(c.Request.Context(), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listRegistryArtifacts(c *gin.Context) {
	data, err := h.modules.ListRegistryArtifacts(c.Request.Context())
	respond(c, data, err)
}

func respond(c *gin.Context, data any, err error) {
	respondWithStatus(c, http.StatusOK, data, err)
}

func respondWithStatus(c *gin.Context, status int, data any, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, data)
}
