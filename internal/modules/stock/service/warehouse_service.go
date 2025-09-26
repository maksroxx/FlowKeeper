package service

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type WarehouseService interface {
	Create(name, address string) (*stock.Warehouse, error)
	GetByID(id uint) (*stock.Warehouse, error)
	List() ([]stock.Warehouse, error)
	Update(w *stock.Warehouse) (*stock.Warehouse, error)
	Delete(id uint) error
}

type warehouseService struct {
	repo repository.WarehouseRepository
}

func NewWarehouseService(r repository.WarehouseRepository) WarehouseService {
	return &warehouseService{repo: r}
}

func (s *warehouseService) Create(name, address string) (*stock.Warehouse, error) {
	return s.repo.Create(&stock.Warehouse{Name: name, Address: address})
}
func (s *warehouseService) GetByID(id uint) (*stock.Warehouse, error) { return s.repo.GetByID(id) }
func (s *warehouseService) List() ([]stock.Warehouse, error)          { return s.repo.List() }
func (s *warehouseService) Update(w *stock.Warehouse) (*stock.Warehouse, error) {
	return s.repo.Update(w)
}
func (s *warehouseService) Delete(id uint) error { return s.repo.Delete(id) }
