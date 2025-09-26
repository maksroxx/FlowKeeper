package service

import (
	"time"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type StockMovementService interface {
	Create(itemID, warehouseID uint, counterpartyID *uint, qty int, mtype, comment string) (*stock.StockMovement, error)
	GetByID(id uint) (*stock.StockMovement, error)
	List() ([]stock.StockMovement, error)
	Update(m *stock.StockMovement) (*stock.StockMovement, error)
	Delete(id uint) error
}

type movementService struct {
	repo repository.StockMovementRepository
}

func NewStockMovementService(r repository.StockMovementRepository) StockMovementService {
	return &movementService{repo: r}
}

func (s *movementService) Create(itemID, warehouseID uint, counterpartyID *uint, qty int, mtype, comment string) (*stock.StockMovement, error) {
	return s.repo.Create(&stock.StockMovement{
		ItemID:         itemID,
		WarehouseID:    warehouseID,
		CounterpartyID: counterpartyID,
		Quantity:       qty,
		Type:           mtype,
		Comment:        comment,
		CreatedAt:      time.Now(),
	})
}
func (s *movementService) GetByID(id uint) (*stock.StockMovement, error) { return s.repo.GetByID(id) }
func (s *movementService) List() ([]stock.StockMovement, error)          { return s.repo.List() }
func (s *movementService) Update(m *stock.StockMovement) (*stock.StockMovement, error) {
	return s.repo.Update(m)
}
func (s *movementService) Delete(id uint) error { return s.repo.Delete(id) }
