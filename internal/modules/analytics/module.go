package analytics

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string { return "analytics" }

func (m *Module) Migrate(db *gorm.DB) error {
	return nil
}

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	group := r.Group("/api/v1/analytics")
	handler.Register(group)
}
