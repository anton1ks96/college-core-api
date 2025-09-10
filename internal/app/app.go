package app

import (
	"fmt"
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Run() {
	cfg, err := config.InitConfig("./configs")
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	if err := router.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)); err != nil {
		logger.Fatal(err)
	}
}
