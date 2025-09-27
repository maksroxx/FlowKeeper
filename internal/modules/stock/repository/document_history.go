package repository

import (
	"gorm.io/gorm"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type DocumentHistoryRepository interface {
	Add(history *stock.DocumentHistory) error
	CreateWithTx(tx *gorm.DB, history *stock.DocumentHistory) error
	GetByDocumentID(docID uint) ([]stock.DocumentHistory, error)
}

type documentHistoryRepo struct {
	db *gorm.DB
}

func NewDocumentHistoryRepository(db *gorm.DB) DocumentHistoryRepository {
	return &documentHistoryRepo{db: db}
}

func (r *documentHistoryRepo) Add(history *stock.DocumentHistory) error {
	return r.db.Create(history).Error
}

func (r *documentHistoryRepo) CreateWithTx(tx *gorm.DB, history *stock.DocumentHistory) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(history).Error
}

func (r *documentHistoryRepo) GetByDocumentID(docID uint) ([]stock.DocumentHistory, error) {
	var records []stock.DocumentHistory
	if err := r.db.Where("document_id = ?", docID).Order("created_at asc").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
