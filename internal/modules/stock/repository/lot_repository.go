package repository

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LotRepository interface {
	GetOldestLotsForUpdate(tx *gorm.DB, warehouseID, variantID uint) ([]models.StockLot, error)
	CreateWithTx(tx *gorm.DB, lot *models.StockLot) error
	SaveWithTx(tx *gorm.DB, lot *models.StockLot) error
	DeleteWithTx(tx *gorm.DB, lotIDs []uint) error
	DeleteByIncomeDocumentID(tx *gorm.DB, docID uint) error
}

type lotRepo struct{}

func NewLotRepository() LotRepository {
	return &lotRepo{}
}

func (r *lotRepo) GetOldestLotsForUpdate(tx *gorm.DB, warehouseID, variantID uint) ([]models.StockLot, error) {
	var lots []models.StockLot
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND variant_id = ? AND current_quantity > 0", warehouseID, variantID).
		Order("arrival_date asc, id asc").
		Find(&lots).Error
	return lots, err
}

func (r *lotRepo) CreateWithTx(tx *gorm.DB, lot *models.StockLot) error {
	return tx.Create(lot).Error
}

func (r *lotRepo) SaveWithTx(tx *gorm.DB, lot *models.StockLot) error {
	return tx.Save(lot).Error
}

func (r *lotRepo) DeleteWithTx(tx *gorm.DB, lotIDs []uint) error {
	if len(lotIDs) == 0 {
		return nil
	}
	return tx.Where("id IN ?", lotIDs).Delete(&models.StockLot{}).Error
}

func (r *lotRepo) DeleteByIncomeDocumentID(tx *gorm.DB, docID uint) error {
	return tx.Where("income_document_id = ?", docID).Delete(&models.StockLot{}).Error
}
