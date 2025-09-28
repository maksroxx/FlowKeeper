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

	GetAvailableQuantity(warehouseID, variantID uint) (decimal.Decimal, error)
	GetBalance(warehouseID, variantID uint) (*stock.StockBalance, error)
	ListByWarehouse(warehouseID uint) ([]stock.StockBalance, error)
	ListByWarehouseFiltered(warehouseID uint, f stock.StockFilter) ([]stock.StockBalance, error)
}

type inventoryService struct {
	balanceRepo     repository.BalanceRepository
	reservationRepo repository.ReservationRepository
	movementRepo    repository.StockMovementRepository
}

func NewInventoryService(b repository.BalanceRepository, r repository.ReservationRepository, m repository.StockMovementRepository) InventoryService {
	return &inventoryService{
		balanceRepo:     b,
		reservationRepo: r,
		movementRepo:    m,
	}
}

func (s *inventoryService) GetAvailableQuantity(warehouseID, variantID uint) (decimal.Decimal, error) {
	return s.getAvailableQuantityWithTx(nil, warehouseID, variantID)
}

func (s *inventoryService) GetBalance(warehouseID, variantID uint) (*stock.StockBalance, error) {
	return s.balanceRepo.GetBalanceWithTx(nil, warehouseID, variantID)
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
	case "ORDER":
		return s.processOrderWithTx(tx, doc)
	case "TRANSFER":
		return s.processTransferWithTx(tx, doc)
	case "INVENTORY":
		return s.processInventoryWithTx(tx, doc)
	default:
		return fmt.Errorf("unsupported document type for inventory processing: %s", doc.Type)
	}
}

func (s *inventoryService) RevertDocumentWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc == nil {
		return errors.New("nil document")
	}

	switch toUpper(doc.Type) {
	case "INCOME", "OUTCOME", "TRANSFER", "INVENTORY":
		return s.revertPhysicalMovement(tx, doc)
	case "ORDER":
		return s.revertOrder(tx, doc)
	default:
		return fmt.Errorf("revert for document type '%s' is not implemented", doc.Type)
	}
}

func (s *inventoryService) getAvailableQuantityWithTx(tx *gorm.DB, warehouseID, variantID uint) (decimal.Decimal, error) {
	balance, err := s.balanceRepo.GetBalanceWithTx(tx, warehouseID, variantID)
	if err != nil {
		return decimal.Zero, err
	}
	reservation, err := s.reservationRepo.GetReservationWithTx(tx, warehouseID, variantID)
	if err != nil {
		return decimal.Zero, err
	}

	balanceQty := decimal.Zero
	if balance != nil {
		balanceQty = balance.Quantity
	}
	reservationQty := decimal.Zero
	if reservation != nil {
		reservationQty = reservation.Quantity
	}

	return balanceQty.Sub(reservationQty), nil
}

func (s *inventoryService) processOrderWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required for order")
	}
	for _, item := range doc.Items {
		available, err := s.getAvailableQuantityWithTx(tx, *doc.WarehouseID, item.VariantID)
		if err != nil {
			return err
		}

		if available.LessThan(item.Quantity) {
			return fmt.Errorf("not enough available stock for variant %d: available=%s, needed=%s", item.VariantID, available.String(), item.Quantity.String())
		}

		res, err := s.reservationRepo.GetReservationWithTx(tx, *doc.WarehouseID, item.VariantID)
		if err != nil {
			return err
		}
		if res == nil {
			res = &stock.StockReservation{
				WarehouseID: *doc.WarehouseID,
				VariantID:   item.VariantID,
				Quantity:    decimal.Zero,
			}
		}
		res.Quantity = res.Quantity.Add(item.Quantity)
		if err := s.reservationRepo.SaveReservationWithTx(tx, res); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processOutcomeWithTx(tx *gorm.DB, doc *stock.Document) error {
	if err := s.processPhysicalOutcomeWithTx(tx, doc); err != nil {
		return err
	}

	if doc.BaseDocumentID != nil {
		if doc.WarehouseID == nil {
			return errors.New("warehouse_id required for outcome based on order")
		}
		for _, item := range doc.Items {
			res, err := s.reservationRepo.GetReservationWithTx(tx, *doc.WarehouseID, item.VariantID)
			if err != nil {
				return err
			}
			if res == nil {
				continue
			}

			res.Quantity = res.Quantity.Sub(item.Quantity)
			if err := s.reservationRepo.SaveReservationWithTx(tx, res); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *inventoryService) revertOrder(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required to revert order")
	}
	for _, item := range doc.Items {
		res, err := s.reservationRepo.GetReservationWithTx(tx, *doc.WarehouseID, item.VariantID)
		if err != nil {
			return err
		}
		if res == nil {
			continue
		}

		res.Quantity = res.Quantity.Sub(item.Quantity)
		if err := s.reservationRepo.SaveReservationWithTx(tx, res); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processIncomeWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required for income")
	}
	for _, it := range doc.Items {
		if it.Price == nil {
			return fmt.Errorf("price must be specified for income item ID %d", it.VariantID)
		}

		itemCost := *it.Price
		mv := &stock.StockMovement{
			DocumentID: &doc.ID, VariantID: it.VariantID, WarehouseID: *doc.WarehouseID,
			CounterpartyID: doc.CounterpartyID, Quantity: it.Quantity, Cost: itemCost,
			Type: "INCOME", Comment: doc.Comment, CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.VariantID)
		if err != nil {
			return err
		}
		if bal == nil {
			bal = &stock.StockBalance{
				WarehouseID: *doc.WarehouseID, VariantID: it.VariantID,
				Quantity: decimal.Zero, TotalCost: decimal.Zero,
			}
		}

		bal.Quantity = bal.Quantity.Add(it.Quantity)
		bal.TotalCost = bal.TotalCost.Add(it.Quantity.Mul(itemCost))

		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processPhysicalOutcomeWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required for outcome")
	}
	for _, it := range doc.Items {
		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.VariantID)
		if err != nil || bal == nil {
			return fmt.Errorf("no stock for variant %d", it.VariantID)
		}
		if bal.Quantity.LessThan(it.Quantity) {
			return fmt.Errorf("not enough stock for variant %d: have=%s need=%s", it.VariantID, bal.Quantity.String(), it.Quantity.String())
		}

		var averageCost decimal.Decimal
		if !bal.Quantity.IsZero() {
			averageCost = bal.TotalCost.Div(bal.Quantity).Round(4)
		}

		mv := &stock.StockMovement{
			DocumentID: &doc.ID, VariantID: it.VariantID, WarehouseID: *doc.WarehouseID,
			CounterpartyID: doc.CounterpartyID, Quantity: it.Quantity.Neg(), Cost: averageCost,
			Type: "OUTCOME", Comment: doc.Comment, CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal.Quantity = bal.Quantity.Sub(it.Quantity)
		bal.TotalCost = bal.TotalCost.Sub(it.Quantity.Mul(averageCost))

		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) revertPhysicalMovement(tx *gorm.DB, doc *stock.Document) error {
	moves, err := s.movementRepo.ListByDocumentWithTx(tx, doc.ID)
	if err != nil {
		return err
	}
	if len(moves) == 0 {
		return nil
	}

	for _, mv := range moves {
		cancel := &stock.StockMovement{
			DocumentID: &doc.ID, VariantID: mv.VariantID, WarehouseID: mv.WarehouseID,
			Quantity: mv.Quantity.Neg(), Cost: mv.Cost.Neg(),
			Type: "CANCEL", Comment: "cancel of doc " + doc.Number, CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, cancel); err != nil {
			return err
		}

		bal, err := s.balanceRepo.GetBalanceWithTx(tx, mv.WarehouseID, mv.VariantID)
		if err != nil {
			return err
		}
		if bal == nil {
			bal = &stock.StockBalance{WarehouseID: mv.WarehouseID, VariantID: mv.VariantID}
		}

		bal.Quantity = bal.Quantity.Sub(mv.Quantity)
		bal.TotalCost = bal.TotalCost.Sub(mv.Quantity.Mul(mv.Cost))

		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processTransferWithTx(tx *gorm.DB, doc *stock.Document) error {
	return errors.New("transfer with cost/reservation calculation is not yet implemented")
}
func (s *inventoryService) processInventoryWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required for inventory")
	}

	for _, it := range doc.Items {
		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.VariantID)
		if err != nil {
			return err
		}

		if bal == nil {
			bal = &stock.StockBalance{
				WarehouseID: *doc.WarehouseID,
				VariantID:   it.VariantID,
				Quantity:    decimal.Zero,
				TotalCost:   decimal.Zero,
			}
		}

		currentQty := bal.Quantity
		currentCost := bal.TotalCost
		newQty := it.Quantity
		movementQty := newQty.Sub(currentQty)

		var movementCost decimal.Decimal
		if movementQty.IsPositive() {
			if it.Price != nil {
				movementCost = *it.Price
			} else {
				movementCost = decimal.Zero
			}
		} else {
			if !currentQty.IsZero() {
				movementCost = currentCost.Div(currentQty).Round(4)
			}
		}

		mv := &stock.StockMovement{
			DocumentID: &doc.ID, VariantID: it.VariantID, WarehouseID: *doc.WarehouseID,
			Quantity: movementQty, Cost: movementCost, Type: "INVENTORY",
			Comment: doc.Comment, CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal.Quantity = newQty
		costChange := movementQty.Mul(movementCost)
		bal.TotalCost = bal.TotalCost.Add(costChange)

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
