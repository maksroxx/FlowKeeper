package api

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/core"
	"github.com/maksroxx/flowkeeper/internal/modules/audit"
	"github.com/maksroxx/flowkeeper/internal/modules/users"
)

func InitAPI(r *gin.Engine, app *core.App, auditSvc audit.Service, authCfg config.AuthConfig) {

	api := r.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	protected := api.Group("")
	protected.Use(users.AuthMiddleware(authCfg))
	protected.Use(audit.AutomaticAuditMiddleware(auditSvc))

	for _, module := range app.Modules() {
		module.RegisterRoutes(r, app.DB())
	}
}
