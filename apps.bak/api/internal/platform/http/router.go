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
	overview *application.OverviewService
	modules  *application.ModuleService
}

func NewRouter(cfg config.Config, overview *application.OverviewService, modules *application.ModuleService, authMiddleware *auth.Middleware) *gin.Engine {
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

	h := &Handler{overview: overview, modules: modules}

	router.GET("/health", h.health)

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth())
	{
		v1.GET("/overview", h.getOverview)
		v1.GET("/applications", h.listApplications)
		v1.POST("/applications", h.createApplication)
		v1.GET("/pipelines", h.listPipelines)
		v1.GET("/deployments", h.listDeployments)
		v1.GET("/security/alerts", h.listSecurityAlerts)
		v1.GET("/observability/signals", h.listObservabilitySignals)
		v1.GET("/integrations", h.listIntegrations)
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
	respondWithStatus(c, http.StatusCreated, data, err)
}

func (h *Handler) listPipelines(c *gin.Context) {
	data, err := h.modules.ListPipelines(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listDeployments(c *gin.Context) {
	data, err := h.modules.ListDeployments(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listSecurityAlerts(c *gin.Context) {
	data, err := h.modules.ListSecurityAlerts(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listObservabilitySignals(c *gin.Context) {
	data, err := h.modules.ListObservabilitySignals(c.Request.Context())
	respond(c, data, err)
}

func (h *Handler) listIntegrations(c *gin.Context) {
	data, err := h.modules.ListIntegrations(c.Request.Context())
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
