package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/config"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type FifoQuantityStrategy struct {
	lotRepo      repository.LotRepository
	movementRepo repository.StockMovementRepository
	balanceRepo  repository.BalanceRepository
	variantRepo  repository.VariantRepository
	productRepo  repository.ProductRepository
}

func NewFifoQuantityStrategy(
	l repository.LotRepository,
	m repository.StockMovementRepository,
	b repository.BalanceRepository,
	v repository.VariantRepository,
	p repository.ProductRepository,
) QuantityStrategy {
	return &FifoQuantityStrategy{
		lotRepo:      l,
		movementRepo: m,
		balanceRepo:  b,
		variantRepo:  v,
		productRepo:  p,
	}
}

func (s *FifoQuantityStrategy) ProcessIncome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required")
	}
	for _, it := range doc.Items {
		lot := &models.StockLot{
			WarehouseID: *doc.WarehouseID, VariantID: it.VariantID,
			IncomeDocumentID: doc.ID, ArrivalDate: doc.CreatedAt, CurrentQuantity: it.Quantity,
		}
		if err := s.lotRepo.CreateWithTx(tx, lot); err != nil {
			return err
		}

		mv := &models.StockMovement{
			DocumentID: &doc.ID, VariantID: it.VariantID, WarehouseID: *doc.WarehouseID,
			Quantity: it.Quantity, Type: "INCOME", CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		updateTotalQuantity(tx, s.balanceRepo, *doc.WarehouseID, it.VariantID, it.Quantity)
	}
	return nil
}

func (s *FifoQuantityStrategy) ProcessOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required")
	}
	for _, it := range doc.Items {
		lots, err := s.lotRepo.GetOldestLotsForUpdate(tx, *doc.WarehouseID, it.VariantID)
		if err != nil {
			return err
		}

		totalQtyInLots := decimal.Zero
		for _, lot := range lots {
			totalQtyInLots = totalQtyInLots.Add(lot.CurrentQuantity)
		}

		if !cfg.AllowNegativeStock && totalQtyInLots.LessThan(it.Quantity) {
			variant, _ := s.variantRepo.GetByID(it.VariantID)
			productName := "Unknown"
			sku := "?"
			if variant != nil {
				sku = variant.SKU
				if p, _ := s.productRepo.GetByID(variant.ProductID); p != nil {
					productName = p.Name
				}
			}
			return fmt.Errorf("Недостаточно товара '%s' (%s). В наличии: %s, Нужно: %s",
				productName, sku, totalQtyInLots.String(), it.Quantity.String())
		}

		qtyToShip := it.Quantity
		lotsToDelete := []uint{}
		for i := range lots {
			if qtyToShip.IsZero() {
				break
			}

			lot := &lots[i]
			qtyFromLot := decimal.Min(qtyToShip, lot.CurrentQuantity)

			mv := &models.StockMovement{
				DocumentID: &doc.ID, VariantID: lot.VariantID, WarehouseID: lot.WarehouseID,
				Quantity: qtyFromLot.Neg(), Type: "OUTCOME", SourceLotID: &lot.ID, CreatedAt: time.Now(),
			}
			if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
				return err
			}

			lot.CurrentQuantity = lot.CurrentQuantity.Sub(qtyFromLot)
			if lot.CurrentQuantity.IsZero() {
				lotsToDelete = append(lotsToDelete, lot.ID)
			} else {
				if err := s.lotRepo.SaveWithTx(tx, lot); err != nil {
					return err
				}
			}
			qtyToShip = qtyToShip.Sub(qtyFromLot)
		}

		if err := s.lotRepo.DeleteWithTx(tx, lotsToDelete); err != nil {
			return err
		}
		updateTotalQuantity(tx, s.balanceRepo, *doc.WarehouseID, it.VariantID, it.Quantity.Neg())
	}
	return nil
}

func (s *FifoQuantityStrategy) RevertIncome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	if err := s.revertMovementsAndUpdateBalance(tx, doc); err != nil {
		return err
	}
	return s.lotRepo.DeleteByIncomeDocumentID(tx, doc.ID)
}

func (s *FifoQuantityStrategy) RevertOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	moves, err := s.movementRepo.ListByDocumentWithTx(tx, doc.ID)
	if err != nil {
		return err
	}

	for _, mv := range moves {
		if mv.Quantity.IsPositive() || mv.SourceLotID == nil {
			continue
		}

		lot, err := s.lotRepo.GetLotByIDForUpdate(tx, *mv.SourceLotID)
		if err != nil {
			return fmt.Errorf("source lot with ID %d not found for movement %d", *mv.SourceLotID, mv.ID)
		}

		lot.CurrentQuantity = lot.CurrentQuantity.Sub(mv.Quantity)
		if err := s.lotRepo.SaveWithTx(tx, lot); err != nil {
			return err
		}
	}

	return s.revertMovementsAndUpdateBalance(tx, doc)
}

func updateTotalQuantity(tx *gorm.DB, balanceRepo repository.BalanceRepository, whID, varID uint, qtyChange decimal.Decimal) {
	bal, _ := balanceRepo.GetBalanceWithTx(tx, whID, varID)
	if bal == nil {
		bal = &models.StockBalance{WarehouseID: whID, VariantID: varID, Quantity: decimal.Zero}
	}
	bal.Quantity = bal.Quantity.Add(qtyChange)
	balanceRepo.SaveBalanceWithTx(tx, bal)
}

func (s *FifoQuantityStrategy) revertMovementsAndUpdateBalance(tx *gorm.DB, doc *models.Document) error {
	moves, err := s.movementRepo.ListByDocumentWithTx(tx, doc.ID)
	if err != nil {
		return err
	}

	for _, mv := range moves {
		cancel := &models.StockMovement{
			DocumentID: &doc.ID, VariantID: mv.VariantID, WarehouseID: mv.WarehouseID,
			Quantity: mv.Quantity.Neg(), Type: "CANCEL", CreatedAt: time.Now(), SourceLotID: mv.SourceLotID,
		}
		if _, err := s.movementRepo.CreateWithTx(tx, cancel); err != nil {
			return err
		}

		updateTotalQuantity(tx, s.balanceRepo, mv.WarehouseID, mv.VariantID, mv.Quantity.Neg())
	}
	return nil
}
