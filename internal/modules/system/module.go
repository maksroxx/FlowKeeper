package system

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/config"
	"gorm.io/gorm"
)

type Module struct {
	dbConfig config.DatabaseConfig
}

func NewModule(cfg config.DatabaseConfig) *Module {
	return &Module{dbConfig: cfg}
}

func (m *Module) Name() string { return "system" }

func (m *Module) Migrate(db *gorm.DB) error {
	return nil
}

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	handler := NewHandler(db, m.dbConfig)
	group := r.Group("/api/v1/system")
	handler.Register(group)
}
