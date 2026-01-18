package files

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "files"
}

func (m *Module) Migrate(db *gorm.DB) error {
	return nil
}

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	r.Static("/uploads", "./uploads")

	handler := NewHandler()
	group := r.Group("/api/v1")
	handler.Register(group)
}
