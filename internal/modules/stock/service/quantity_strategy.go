package service

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/config"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type QuantityStrategy interface {
	ProcessIncome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error
	ProcessOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error
	RevertIncome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error
	RevertOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error
}
