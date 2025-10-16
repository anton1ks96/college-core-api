package handlers

import (
	"github.com/anton1ks96/college-core-api/internal/config"
	v1 "github.com/anton1ks96/college-core-api/internal/handlers/v1"
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

func (h *Handler) Init() *gin.Engine {
	router := gin.New()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		httpmw.ErrorHandlerMiddleware(),
		httpmw.CORSMiddleware(),
		httpmw.RequestIDMiddleware(),
	)

	router.GET("/health", h.healthCheck)

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	api := router.Group("/api")

	v1Handler := v1.NewHandler(h.services, h.cfg)
	v1Group := api.Group("/v1")

	v1Group.Use(httpmw.AuthMiddleware(h.services.Auth))

	v1Handler.Init(v1Group)
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "OK",
		"service": "college-core-api",
	})
}
