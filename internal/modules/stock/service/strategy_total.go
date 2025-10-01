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

type TotalQuantityStrategy struct {
	balanceRepo  repository.BalanceRepository
	movementRepo repository.StockMovementRepository
}

func NewTotalQuantityStrategy(b repository.BalanceRepository, m repository.StockMovementRepository) QuantityStrategy {
	return &TotalQuantityStrategy{balanceRepo: b, movementRepo: m}
}

func (s *TotalQuantityStrategy) ProcessIncome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required")
	}
	for _, it := range doc.Items {
		mv := &models.StockMovement{
			DocumentID: &doc.ID, VariantID: it.VariantID, WarehouseID: *doc.WarehouseID,
			Quantity: it.Quantity, Type: "INCOME", CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.VariantID)
		if err != nil {
			return err
		}
		if bal == nil {
			bal = &models.StockBalance{WarehouseID: *doc.WarehouseID, VariantID: it.VariantID, Quantity: decimal.Zero}
		}
		bal.Quantity = bal.Quantity.Add(it.Quantity)
		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *TotalQuantityStrategy) ProcessOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required")
	}
	for _, it := range doc.Items {
		bal, err := s.balanceRepo.GetBalanceWithTx(tx, *doc.WarehouseID, it.VariantID)
		if err != nil {
			return err
		}

		currentQty := decimal.Zero
		if bal != nil {
			currentQty = bal.Quantity
		}

		if !cfg.AllowNegativeStock && currentQty.LessThan(it.Quantity) {
			return fmt.Errorf("not enough stock for variant %d: have=%s, need=%s", it.VariantID, currentQty.String(), it.Quantity.String())
		}

		mv := &models.StockMovement{
			DocumentID: &doc.ID, VariantID: it.VariantID, WarehouseID: *doc.WarehouseID,
			Quantity: it.Quantity.Neg(), Type: "OUTCOME", CreatedAt: time.Now(),
		}
		if _, err := s.movementRepo.CreateWithTx(tx, mv); err != nil {
			return err
		}

		if bal == nil {
			bal = &models.StockBalance{WarehouseID: *doc.WarehouseID, VariantID: it.VariantID, Quantity: decimal.Zero}
		}
		bal.Quantity = bal.Quantity.Sub(it.Quantity)
		if err := s.balanceRepo.SaveBalanceWithTx(tx, bal); err != nil {
			return err
		}
	}
	return nil
}

func (s *TotalQuantityStrategy) RevertIncome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	// ... (здесь логика отмены, которая работает с StockBalance)
	return nil
}

func (s *TotalQuantityStrategy) RevertOutcome(tx *gorm.DB, doc *models.Document, cfg *config.Config) error {
	// ... (здесь логика отмены, которая работает с StockBalance)
	return nil
}
