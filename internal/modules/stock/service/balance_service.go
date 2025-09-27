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
	ProcessDocument(doc *stock.Document) error
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
	return &inventoryService{
		balanceRepo:  b,
		movementRepo: m,
		tx:           tx,
	}
}

func (s *inventoryService) GetBalance(warehouseID, itemID uint) (decimal.Decimal, error) {
	b, err := s.balanceRepo.GetBalance(warehouseID, itemID)
	if err != nil {
		return decimal.Zero, err
	}
	if b == nil {
		return decimal.Zero, nil
	}
	return b.Quantity, nil
}

func (s *inventoryService) ListByWarehouse(warehouseID uint) ([]stock.StockBalance, error) {
	return s.balanceRepo.ListByWarehouse(warehouseID)
}

func (s *inventoryService) ListByWarehouseFiltered(warehouseID uint, f stock.StockFilter) ([]stock.StockBalance, error) {
	return s.balanceRepo.ListByWarehouseFiltered(warehouseID, f)
}

func (s *inventoryService) ProcessDocument(doc *stock.Document) error {
	return s.ProcessDocumentWithTx(nil, doc)
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

func toUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s)
}

func (s *inventoryService) processIncomeWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required for income")
	}
	for _, it := range doc.Items {
		mv := &stock.StockMovement{
			DocumentID:     &doc.ID,
			ItemID:         it.ItemID,
			WarehouseID:    *doc.WarehouseID,
			CounterpartyID: doc.CounterpartyID,
			Quantity:       it.Quantity,
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
			}
		}
		bal.Quantity = bal.Quantity.Add(it.Quantity)
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
		if err != nil {
			return err
		}
		cur := decimal.Zero
		if bal != nil {
			cur = bal.Quantity
		}
		if cur.LessThan(it.Quantity) {
			return fmt.Errorf("not enough stock for item %d on warehouse %d: have=%s need=%s", it.ItemID, *doc.WarehouseID, cur.String(), it.Quantity.String())
		}

		mv := &stock.StockMovement{
			DocumentID:     &doc.ID,
			ItemID:         it.ItemID,
			WarehouseID:    *doc.WarehouseID,
			CounterpartyID: doc.CounterpartyID,
			Quantity:       it.Quantity.Neg(),
			Type:           "OUTCOME",
			Comment:        doc.Comment,
			CreatedAt:      time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal.Quantity = bal.Quantity.Sub(it.Quantity)
		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processTransferWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil || doc.ToWarehouseID == nil {
		return errors.New("both WarehouseID (from) and ToWarehouseID (to) required for transfer")
	}
	from := *doc.WarehouseID
	to := *doc.ToWarehouseID

	for _, it := range doc.Items {
		balFrom, err := s.balanceRepo.GetBalanceWithTx(tx, from, it.ItemID)
		if err != nil {
			return err
		}

		cur := decimal.Zero
		if balFrom != nil {
			cur = balFrom.Quantity
		}

		if cur.LessThan(it.Quantity) {
			return fmt.Errorf("not enough stock for transfer item %d: have=%s need=%s", it.ItemID, cur.String(), it.Quantity.String())
		}

		mvOut := &stock.StockMovement{
			DocumentID:  &doc.ID,
			ItemID:      it.ItemID,
			WarehouseID: from,
			Quantity:    it.Quantity.Neg(),
			Type:        "TRANSFER",
			Comment:     "transfer out: " + doc.Comment,
			CreatedAt:   time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mvOut); err != nil {
			return err
		}

		mvIn := &stock.StockMovement{
			DocumentID:  &doc.ID,
			ItemID:      it.ItemID,
			WarehouseID: to,
			Quantity:    it.Quantity,
			Type:        "TRANSFER",
			Comment:     "transfer in: " + doc.Comment,
			CreatedAt:   time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mvIn); err != nil {
			return err
		}

		balFrom.Quantity = balFrom.Quantity.Sub(it.Quantity)
		if err := s.balanceRepo.SaveBalanceWithTx(tx, balFrom); err != nil {
			return err
		}

		balTo, err := s.balanceRepo.GetBalanceWithTx(tx, to, it.ItemID)
		if err != nil {
			return err
		}
		if balTo == nil {
			balTo = &stock.StockBalance{WarehouseID: to, ItemID: it.ItemID, Quantity: decimal.Zero}
		}
		balTo.Quantity = balTo.Quantity.Add(it.Quantity)
		if err := s.balanceRepo.SaveBalanceWithTx(tx, balTo); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processInventoryWithTx(tx *gorm.DB, doc *stock.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required for inventory")
	}
	for _, it := range doc.Items {
		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.ItemID)
		if err != nil {
			return err
		}
		if bal == nil {
			bal = &stock.StockBalance{
				WarehouseID: *doc.WarehouseID,
				ItemID:      it.ItemID,
			}
		}

		currentQty := bal.Quantity
		movementQty := it.Quantity.Sub(currentQty)

		mv := &stock.StockMovement{
			DocumentID:  &doc.ID,
			ItemID:      it.ItemID,
			WarehouseID: *doc.WarehouseID,
			Quantity:    movementQty,
			Type:        "INVENTORY",
			Comment:     doc.Comment,
			CreatedAt:   time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal.Quantity = it.Quantity
		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
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
			bal = &stock.StockBalance{WarehouseID: mv.WarehouseID, ItemID: mv.ItemID, Quantity: decimal.Zero}
		}
		bal.Quantity = bal.Quantity.Sub(mv.Quantity)
		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}
