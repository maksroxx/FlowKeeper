package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/config"
	"gorm.io/gorm"
)

type Module struct {
	Service Service
}

func NewModule(db *gorm.DB, cfg config.AuditConfig) *Module {
	svc := NewAsyncService(db, cfg)
	return &Module{Service: svc}
}

func (m *Module) Name() string { return "audit" }

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&AuditLog{})
}

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	handler := NewHandler(m.Service)
	group := r.Group("/api/v1/audit")
	handler.Register(group)
}
