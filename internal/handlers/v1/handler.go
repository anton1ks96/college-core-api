package v1

import (
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/httpmw"
	"github.com/anton1ks96/college-core-api/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *services.Services
	cfg      *config.Config
}

func NewHandler(services *services.Services, cfg *config.Config) *Handler {
	return &Handler{
		services: services,
		cfg:      cfg,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	datasets := api.Group("/datasets")
	{
		datasets.POST("", httpmw.RateLimitMiddleware(h.cfg.Limits.UploadRateLimit), h.createDataset)

		datasets.GET("", h.getDatasets)
		datasets.GET("/:id", h.getDataset)
		datasets.PUT("/:id", h.updateDataset)
		datasets.DELETE("/:id", h.deleteDataset)

		datasets.POST("/:id/ask", httpmw.RateLimitMiddleware(h.cfg.Limits.AskRateLimit), h.askQuestion)
		datasets.POST("/:id/reindex", h.reindexDataset)
	}
}
