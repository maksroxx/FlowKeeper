package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type ItemRepository interface {
	Create(i *stock.Item) (*stock.Item, error)
	GetByID(id uint) (*stock.Item, error)
	List() ([]stock.Item, error)
	Update(i *stock.Item) (*stock.Item, error)
	Delete(id uint) error
}

type itemRepo struct{ db *gorm.DB }

func NewItemRepository(db *gorm.DB) ItemRepository { return &itemRepo{db: db} }

func (r *itemRepo) Create(i *stock.Item) (*stock.Item, error) {
	err := r.db.Create(i).Error
	return i, err
}

func (r *itemRepo) GetByID(id uint) (*stock.Item, error) {
	var item stock.Item
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *itemRepo) List() ([]stock.Item, error) {
	var items []stock.Item
	err := r.db.Find(&items).Error
	return items, err
}

func (r *itemRepo) Update(i *stock.Item) (*stock.Item, error) {
	err := r.db.Save(i).Error
	return i, err
}

func (r *itemRepo) Delete(id uint) error {
	return r.db.Delete(&stock.Item{}, id).Error
}
