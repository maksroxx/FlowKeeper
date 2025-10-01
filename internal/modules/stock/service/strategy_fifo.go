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
}

func NewFifoQuantityStrategy(l repository.LotRepository, m repository.StockMovementRepository, b repository.BalanceRepository) QuantityStrategy {
	return &FifoQuantityStrategy{lotRepo: l, movementRepo: m, balanceRepo: b}
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
			return fmt.Errorf("not enough stock for variant %d: have=%s, need=%s", it.VariantID, totalQtyInLots.String(), it.Quantity.String())
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
				Quantity: qtyFromLot.Neg(), Type: "OUTCOME", CreatedAt: time.Now(),
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
	// ... логика отмены: удалить партии и движения, обновить StockBalance
	return nil
}
func (s *FifoQuantityStrategy) RevertOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	// ... более сложная логика: "вернуть" количество в нужные партии
	return nil
}

func updateTotalQuantity(tx *gorm.DB, balanceRepo repository.BalanceRepository, whID, varID uint, qtyChange decimal.Decimal) {
	bal, _ := balanceRepo.GetBalanceWithTx(tx, whID, varID)
	if bal == nil {
		bal = &models.StockBalance{WarehouseID: whID, VariantID: varID, Quantity: decimal.Zero}
	}
	bal.Quantity = bal.Quantity.Add(qtyChange)
	balanceRepo.SaveBalanceWithTx(tx, bal)
}
