package shop

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/modules/users"
	"gorm.io/gorm"
)

type Module struct {
	authConfig config.AuthConfig
}

func NewModule(authCfg config.AuthConfig) *Module {
	return &Module{authConfig: authCfg}
}

func (m *Module) Name() string { return "shop" }

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ShopSettings{})
}

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	repo := NewRepository(db)
	handler := NewHandler(repo)

	shopGroup := r.Group("/api/v1/shop")

	shopGroup.GET("/catalog", handler.GetCatalog)
	shopGroup.GET("/product/:id", handler.GetProduct)
	shopGroup.GET("/settings", handler.GetSettings)

	adminGroup := shopGroup.Group("/admin")
	adminGroup.Use(users.AuthMiddleware(m.authConfig))
	{
		adminGroup.PUT("/settings", handler.UpdateSettings)
		adminGroup.POST("/product/:id/status", handler.ToggleStatus)
	}
}
