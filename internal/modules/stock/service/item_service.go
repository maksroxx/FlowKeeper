package service

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type ItemService interface {
	Create(name, sku string, unitID, categoryID uint, price float64) (*stock.Item, error)
	GetByID(id uint) (*stock.Item, error)
	List() ([]stock.Item, error)
	Update(i *stock.Item) (*stock.Item, error)
	Delete(id uint) error
}

type itemService struct{ repo repository.ItemRepository }

func NewItemService(r repository.ItemRepository) ItemService { return &itemService{repo: r} }

func (s *itemService) Create(name, sku string, unitID, categoryID uint, price float64) (*stock.Item, error) {
	return s.repo.Create(&stock.Item{Name: name, SKU: sku, UnitID: unitID, CategoryID: categoryID, Price: price})
}
func (s *itemService) GetByID(id uint) (*stock.Item, error)      { return s.repo.GetByID(id) }
func (s *itemService) List() ([]stock.Item, error)               { return s.repo.List() }
func (s *itemService) Update(i *stock.Item) (*stock.Item, error) { return s.repo.Update(i) }
func (s *itemService) Delete(id uint) error                      { return s.repo.Delete(id) }
