package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type WarehouseRepository interface {
	Create(w *stock.Warehouse) (*stock.Warehouse, error)
	GetByID(id uint) (*stock.Warehouse, error)
	List() ([]stock.Warehouse, error)
	Update(w *stock.Warehouse) (*stock.Warehouse, error)
	Delete(id uint) error
}

type warehouseRepo struct{ db *gorm.DB }

func NewWarehouseRepository(db *gorm.DB) WarehouseRepository { return &warehouseRepo{db: db} }

func (r *warehouseRepo) Create(w *stock.Warehouse) (*stock.Warehouse, error) {
	err := r.db.Create(w).Error
	return w, err
}

func (r *warehouseRepo) GetByID(id uint) (*stock.Warehouse, error) {
	var wh stock.Warehouse
	if err := r.db.First(&wh, id).Error; err != nil {
		return nil, err
	}
	return &wh, nil
}

func (r *warehouseRepo) List() ([]stock.Warehouse, error) {
	var ws []stock.Warehouse
	err := r.db.Find(&ws).Error
	return ws, err
}

func (r *warehouseRepo) Update(w *stock.Warehouse) (*stock.Warehouse, error) {
	err := r.db.Save(w).Error
	return w, err
}

func (r *warehouseRepo) Delete(id uint) error {
	return r.db.Delete(&stock.Warehouse{}, id).Error
}
