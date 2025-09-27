package repository

import (
	"gorm.io/gorm"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type StockMovementRepository interface {
	Create(m *stock.StockMovement) (*stock.StockMovement, error)
	CreateWithTx(tx *gorm.DB, m *stock.StockMovement) (*stock.StockMovement, error)
	List() ([]stock.StockMovement, error)
	GetByID(id uint) (*stock.StockMovement, error)
	Update(m *stock.StockMovement) (*stock.StockMovement, error)
	Delete(id uint) error

	ListByDocument(docID uint) ([]stock.StockMovement, error)
	ListByDocumentWithTx(tx *gorm.DB, docID uint) ([]stock.StockMovement, error)
}

type movementRepo struct {
	db *gorm.DB
}

func NewStockMovementRepository(db *gorm.DB) StockMovementRepository {
	return &movementRepo{db: db}
}

func (r *movementRepo) Create(m *stock.StockMovement) (*stock.StockMovement, error) {
	if err := r.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *movementRepo) CreateWithTx(tx *gorm.DB, m *stock.StockMovement) (*stock.StockMovement, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *movementRepo) List() ([]stock.StockMovement, error) {
	var ms []stock.StockMovement
	if err := r.db.Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

func (r *movementRepo) GetByID(id uint) (*stock.StockMovement, error) {
	var m stock.StockMovement
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *movementRepo) Update(m *stock.StockMovement) (*stock.StockMovement, error) {
	if err := r.db.Save(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *movementRepo) Delete(id uint) error {
	return r.db.Delete(&stock.StockMovement{}, id).Error
}

func (r *movementRepo) ListByDocument(docID uint) ([]stock.StockMovement, error) {
	return r.ListByDocumentWithTx(nil, docID)
}

func (r *movementRepo) ListByDocumentWithTx(tx *gorm.DB, docID uint) ([]stock.StockMovement, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var ms []stock.StockMovement
	if err := db.Where("document_id = ?", docID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}
