package api

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/core"
)

func InitAPI(r *gin.Engine, app *core.App) {
	api := r.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	for _, module := range app.Modules() {
		module.RegisterRoutes(r, app.DB())
	}
}
