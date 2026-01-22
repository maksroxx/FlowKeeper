package users

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/config"
	"gorm.io/gorm"
)

type Module struct {
	config config.AuthConfig
}

func NewModule(cfg config.AuthConfig) *Module {
	return &Module{config: cfg}
}

func (m *Module) Name() string { return "users" }

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Role{}, &User{})
}

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	repo := NewRepository(db)
	authSvc := NewAuthService(repo, m.config)
	userSvc := NewUserService(repo)

	h := NewHandler(authSvc, userSvc)

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", h.Login)
	}

	api := r.Group("/api/v1")
	api.Use(AuthMiddleware(m.config))
	{
		api.GET("/auth/me", h.GetMe)

		// Пользователи (для админов)
		userGroup := api.Group("/users")
		userGroup.Use(RequirePermission("manage_users"))
		{
			userGroup.GET("", h.ListUsers)
			userGroup.POST("", h.CreateUser)
			userGroup.PUT("/:id", h.UpdateUser)
			userGroup.DELETE("/:id", h.DeleteUser)
		}

		// Роли (для админов)
		roleGroup := api.Group("/roles")
		roleGroup.Use(RequirePermission("manage_users"))
		{
			roleGroup.GET("", h.ListRoles)
			roleGroup.POST("", h.CreateRole)
			roleGroup.PUT("/:id", h.UpdateRole)
			roleGroup.DELETE("/:id", h.DeleteRole)
		}
	}
}
