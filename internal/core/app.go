package core

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module interface {
	Name() string
	Migrate(db *gorm.DB) error
	RegisterRoutes(r *gin.Engine, db *gorm.DB)
}

type App struct {
	db      *gorm.DB
	r       *gin.Engine
	modules []Module
}

func NewApp(db *gorm.DB, r *gin.Engine) *App {
	return &App{db: db, r: r}
}

func (a *App) RegisterModule(m Module) {
	a.modules = append(a.modules, m)
}

func (a *App) Migrate() error {
	for _, m := range a.modules {
		if err := m.Migrate(a.db); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Modules() []Module {
	return a.modules
}

func (a *App) DB() *gorm.DB {
	return a.db
}
