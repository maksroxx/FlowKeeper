package stock

import (
	"gorm.io/gorm"
)

type ItemRepository interface {
	Create(item *Item) error
	List() ([]Item, error)
}

type itemRepository struct {
	db *gorm.DB
}

func (r *itemRepository) Create(item *Item) error {
	return r.db.Create(item).Error
}

func (r *itemRepository) List() ([]Item, error) {
	var items []Item
	err := r.db.Find(&items).Error
	return items, err
}
