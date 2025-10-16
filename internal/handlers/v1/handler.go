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

		datasets.GET("/:id/permissions", httpmw.RequireRole("admin"), h.getDatasetPermissions)
		datasets.POST("/:id/permissions", httpmw.RequireRole("admin"), h.grantDatasetPermission)
		datasets.DELETE("/:id/permissions/:teacher_id", httpmw.RequireRole("admin"), h.revokeDatasetPermission)
	}

	permissions := api.Group("/permissions")
	{
		permissions.GET("", httpmw.RequireRole("admin"), h.getAllPermissions)
	}

	topics := api.Group("/topics")
	{
		topics.POST("", httpmw.RequireRole("teacher", "admin"), h.createTopic)
		topics.GET("", httpmw.RequireRole("teacher", "admin"), h.getMyTopics)
		topics.GET("/all", httpmw.RequireRole("admin"), h.getAllTopics)
		topics.POST("/:id/students", httpmw.RequireRole("teacher", "admin"), h.addStudentsToTopic)
		topics.GET("/:id/students", httpmw.RequireRole("teacher", "admin"), h.getTopicStudents)
		topics.DELETE("/:id/students/:student_id", httpmw.RequireRole("teacher", "admin"), h.removeStudentFromTopic)

		topics.GET("/assigned", h.getAssignedTopics)
	}

	search := api.Group("/search")
	{
		search.POST("/students", httpmw.RequireRole("teacher", "admin"), h.searchStudents)
		search.POST("/teachers", httpmw.RequireRole("admin"), h.searchTeachers)
	}
}
