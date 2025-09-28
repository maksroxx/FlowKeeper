package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type InventoryService interface {
	ProcessDocumentWithTx(tx *gorm.DB, doc *stock.Document) error
	RevertDocumentWithTx(tx *gorm.DB, doc *stock.Document) error
	GetBalance(warehouseID, itemID uint) (decimal.Decimal, error)
	ListByWarehouse(warehouseID uint) ([]stock.StockBalance, error)
	ListByWarehouseFiltered(warehouseID uint, f stock.StockFilter) ([]stock.StockBalance, error)
}

type inventoryService struct {
	balanceRepo  repository.BalanceRepository
	movementRepo repository.StockMovementRepository
	tx           repository.TxManager
}

func NewInventoryService(b repository.BalanceRepository, m repository.StockMovementRepository, tx repository.TxManager) InventoryService {
	return &inventoryService{balanceRepo: b, movementRepo: m, tx: tx}
}

func (s *inventoryService) GetBalance(warehouseID, itemID uint) (decimal.Decimal, error) {
	b, err := s.balanceRepo.GetBalance(warehouseID, itemID)
	if err != nil || b == nil {
		return decimal.Zero, err
	}
	return b.Quantity, nil
}

func (s *inventoryService) ListByWarehouse(warehouseID uint) ([]stock.StockBalance, error) {
	return s.balanceRepo.ListByWarehouse(warehouseID)
}

func (s *inventoryService) ListByWarehouseFiltered(warehouseID uint, f stock.StockFilter) ([]stock.StockBalance, error) {
	return s.balanceRepo.ListByWarehouseFiltered(warehouseID, f)
}

func (s *inventoryService) ProcessDocumentWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc == nil {
		return errors.New("nil document")
	}
	if len(doc.Items) == 0 {
		return errors.New("document must contain at least one item")
	}

	switch toUpper(doc.Type) {
	case "INCOME":
		return s.processIncomeWithTx(tx, doc)
	case "OUTCOME":
		return s.processOutcomeWithTx(tx, doc)
	case "TRANSFER":
		return s.processTransferWithTx(tx, doc)
	case "INVENTORY":
		return s.processInventoryWithTx(tx, doc)
	default:
		return fmt.Errorf("unsupported document type: %s", doc.Type)
	}
}

func (s *inventoryService) processIncomeWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required for income")
	}
	for _, it := range doc.Items {
		if it.Price == nil {
			return fmt.Errorf("price must be specified for income item ID %d", it.ItemID)
		}

		itemCost := *it.Price
		mv := &stock.StockMovement{
			DocumentID:     &doc.ID,
			ItemID:         it.ItemID,
			WarehouseID:    *doc.WarehouseID,
			CounterpartyID: doc.CounterpartyID,
			Quantity:       it.Quantity,
			Cost:           itemCost,
			Type:           "INCOME",
			Comment:        doc.Comment,
			CreatedAt:      time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.ItemID)
		if err != nil {
			return err
		}
		if bal == nil {
			bal = &stock.StockBalance{
				WarehouseID: *doc.WarehouseID,
				ItemID:      it.ItemID,
				Quantity:    decimal.Zero,
				TotalCost:   decimal.Zero,
			}
		}

		bal.Quantity = bal.Quantity.Add(it.Quantity)
		lineTotalCost := it.Quantity.Mul(itemCost)
		bal.TotalCost = bal.TotalCost.Add(lineTotalCost)

		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processOutcomeWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required for outcome")
	}
	for _, it := range doc.Items {
		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.ItemID)
		if err != nil || bal == nil {
			return fmt.Errorf("no stock for item %d", it.ItemID)
		}
		if bal.Quantity.LessThan(it.Quantity) {
			return fmt.Errorf("not enough stock for item %d: have=%s need=%s", it.ItemID, bal.Quantity.String(), it.Quantity.String())
		}

		var averageCost decimal.Decimal
		if !bal.Quantity.IsZero() {
			averageCost = bal.TotalCost.Div(bal.Quantity).Round(4)
		}

		mv := &stock.StockMovement{
			DocumentID:     &doc.ID,
			ItemID:         it.ItemID,
			WarehouseID:    *doc.WarehouseID,
			CounterpartyID: doc.CounterpartyID,
			Quantity:       it.Quantity.Neg(),
			Cost:           averageCost,
			Type:           "OUTCOME",
			Comment:        doc.Comment,
			CreatedAt:      time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal.Quantity = bal.Quantity.Sub(it.Quantity)
		lineTotalCost := it.Quantity.Mul(averageCost)
		bal.TotalCost = bal.TotalCost.Sub(lineTotalCost)

		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processTransferWithTx(tx *gorm.DB, doc *stock.Document) error {
	// Эта логика более сложная, так как себестоимость должна "переехать" со склада на склад
	return errors.New("transfer with cost calculation is not yet implemented")
}

func (s *inventoryService) processInventoryWithTx(tx *gorm.DB, doc *stock.Document) error {
	// Эта логика требует решения, как устанавливать себестоимость при инвентаризации
	return errors.New("inventory with cost calculation is not yet implemented")
}

func (s *inventoryService) RevertDocumentWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc == nil {
		return errors.New("document is nil")
	}
	moves, err := s.movementRepo.ListByDocumentWithTx(tx, doc.ID)
	if err != nil {
		return err
	}
	if len(moves) == 0 {
		return nil
	}

	for _, mv := range moves {
		cancel := &stock.StockMovement{
			DocumentID:  &doc.ID,
			ItemID:      mv.ItemID,
			WarehouseID: mv.WarehouseID,
			Quantity:    mv.Quantity.Neg(),
			Cost:        mv.Cost.Neg(),
			Type:        "CANCEL",
			Comment:     "cancel of doc " + doc.Number,
			CreatedAt:   time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, cancel); err != nil {
			return err
		}

		bal, err := s.balanceRepo.GetBalanceWithTx(tx, mv.WarehouseID, mv.ItemID)
		if err != nil {
			return err
		}
		if bal == nil {
			bal = &stock.StockBalance{WarehouseID: mv.WarehouseID, ItemID: mv.ItemID}
		}

		bal.Quantity = bal.Quantity.Sub(mv.Quantity)
		lineTotalCost := mv.Quantity.Mul(mv.Cost)
		bal.TotalCost = bal.TotalCost.Sub(lineTotalCost)

		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func toUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s)
}
