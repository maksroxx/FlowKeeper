package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type StockMovementRepository interface {
	Create(m *stock.StockMovement) (*stock.StockMovement, error)
	GetByID(id uint) (*stock.StockMovement, error)
	List() ([]stock.StockMovement, error)
	Update(m *stock.StockMovement) (*stock.StockMovement, error)
	Delete(id uint) error
}

type movementRepo struct{ db *gorm.DB }

func NewStockMovementRepository(db *gorm.DB) StockMovementRepository { return &movementRepo{db: db} }

func (r *movementRepo) Create(m *stock.StockMovement) (*stock.StockMovement, error) {
	err := r.db.Create(m).Error
	return m, err
}

func (r *movementRepo) GetByID(id uint) (*stock.StockMovement, error) {
	var mov stock.StockMovement
	if err := r.db.First(&mov, id).Error; err != nil {
		return nil, err
	}
	return &mov, nil
}

func (r *movementRepo) List() ([]stock.StockMovement, error) {
	var moves []stock.StockMovement
	err := r.db.Find(&moves).Error
	return moves, err
}

func (r *movementRepo) Update(m *stock.StockMovement) (*stock.StockMovement, error) {
	err := r.db.Save(m).Error
	return m, err
}

func (r *movementRepo) Delete(id uint) error {
	return r.db.Delete(&stock.StockMovement{}, id).Error
}
