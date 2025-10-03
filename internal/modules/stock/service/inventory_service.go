package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/config"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/shopspring/decimal"
)

type StrategyFactory interface {
	GetStrategy(policy string) (QuantityStrategy, error)
}
type strategyFactoryImpl struct {
	balanceRepo  repository.BalanceRepository
	movementRepo repository.StockMovementRepository
	lotRepo      repository.LotRepository
}

func NewStrategyFactory(b repository.BalanceRepository, m repository.StockMovementRepository, l repository.LotRepository) StrategyFactory {
	return &strategyFactoryImpl{balanceRepo: b, movementRepo: m, lotRepo: l}
}
func (f *strategyFactoryImpl) GetStrategy(policy string) (QuantityStrategy, error) {
	switch policy {
	case "total":
		return NewTotalQuantityStrategy(f.balanceRepo, f.movementRepo), nil
	case "fifo":
		return NewFifoQuantityStrategy(f.lotRepo, f.movementRepo, f.balanceRepo), nil
	default:
		return nil, fmt.Errorf("unknown quantity accounting policy: %s", policy)
	}
}

type InventoryService interface {
	ProcessDocumentWithTx(tx *gorm.DB, doc *models.Document) error
	RevertDocumentWithTx(tx *gorm.DB, doc *models.Document) error
	GetAvailableQuantity(warehouseID, variantID uint) (decimal.Decimal, error)
	ListByWarehouseFilteredAsDTO(warehouseID uint, f models.StockFilter) ([]models.StockBalanceDTO, error)
}
type inventoryService struct {
	strategyFactory StrategyFactory
	reservationRepo repository.ReservationRepository
	balanceRepo     repository.BalanceRepository
	config          *config.Config

	variantRepo repository.VariantRepository
	productRepo repository.ProductRepository
	unitRepo    repository.UnitRepository
}

func NewInventoryService(
	factory StrategyFactory,
	r repository.ReservationRepository,
	b repository.BalanceRepository,
	cfg *config.Config,
	v repository.VariantRepository,
	p repository.ProductRepository,
	u repository.UnitRepository,
) InventoryService {
	return &inventoryService{
		strategyFactory: factory,
		reservationRepo: r,
		balanceRepo:     b,
		config:          cfg,
		variantRepo:     v,
		productRepo:     p,
		unitRepo:        u,
	}
}

func (s *inventoryService) ProcessDocumentWithTx(tx *gorm.DB, doc *models.Document) error {
	strategy, err := s.strategyFactory.GetStrategy(s.config.AccountingPolicy)
	if err != nil {
		return err
	}

	switch toUpper(doc.Type) {
	case "ORDER":
		return s.processOrder(tx, doc)
	case "OUTCOME":
		if doc.BaseDocumentID != nil {
			if err := s.processReservationRelease(tx, doc); err != nil {
				return err
			}
		}
		return strategy.ProcessOutcome(tx, doc, s.config)
	case "INCOME":
		return strategy.ProcessIncome(tx, doc, s.config)
	default:
		return fmt.Errorf("document type '%s' not supported for inventory processing", doc.Type)
	}
}

func (s *inventoryService) RevertDocumentWithTx(tx *gorm.DB, doc *models.Document) error {
	policy := s.config.AccountingPolicy
	strategy, err := s.strategyFactory.GetStrategy(policy)
	if err != nil {
		return err
	}

	switch toUpper(doc.Type) {
	case "ORDER":
		return s.revertOrder(tx, doc)
	case "OUTCOME":
		if doc.BaseDocumentID != nil {
			if err := s.revertReservationRelease(tx, doc); err != nil {
				return err
			}
		}
	}

	switch toUpper(doc.Type) {
	case "INCOME":
		return strategy.RevertIncome(tx, doc, s.config)
	case "OUTCOME":
		return strategy.RevertOutcome(tx, doc, s.config)
	}

	return nil
}

func (s *inventoryService) GetAvailableQuantity(warehouseID, variantID uint) (decimal.Decimal, error) {
	balance, err := s.balanceRepo.GetBalanceWithTx(nil, warehouseID, variantID)
	if err != nil {
		return decimal.Zero, err
	}
	reservation, err := s.reservationRepo.GetReservationWithTx(nil, warehouseID, variantID)
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

func (s *inventoryService) ListByWarehouseFilteredAsDTO(warehouseID uint, f models.StockFilter) ([]models.StockBalanceDTO, error) {
	balances, err := s.balanceRepo.ListByWarehouseFiltered(warehouseID, f)
	if err != nil {
		return nil, err
	}
	if len(balances) == 0 {
		return []models.StockBalanceDTO{}, nil
	}

	variantIDs := make([]uint, len(balances))
	for i, b := range balances {
		variantIDs[i] = b.VariantID
	}

	variants, _ := s.variantRepo.GetByIDs(variantIDs)
	variantMap := make(map[uint]models.Variant)
	productIDsMap := make(map[uint]bool)
	unitIDsMap := make(map[uint]bool)
	for _, v := range variants {
		variantMap[v.ID] = v
		productIDsMap[v.ProductID] = true
		unitIDsMap[v.UnitID] = true
	}

	products, _ := s.productRepo.GetByIDs(mapKeysToSlice(productIDsMap))
	productMap := make(map[uint]models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	units, _ := s.unitRepo.GetByIDs(mapKeysToSlice(unitIDsMap))
	unitMap := make(map[uint]string)
	for _, u := range units {
		unitMap[u.ID] = u.Name
	}

	dtos := make([]models.StockBalanceDTO, len(balances))
	for i, b := range balances {
		variant := variantMap[b.VariantID]
		product := productMap[variant.ProductID]
		dtos[i] = models.StockBalanceDTO{
			ID:          b.ID,
			WarehouseID: b.WarehouseID,
			VariantID:   b.VariantID,
			VariantSKU:  variant.SKU,
			ProductName: product.Name,
			UnitName:    unitMap[variant.UnitID],
			Quantity:    b.Quantity,
		}
	}

	return dtos, nil
}

func (s *inventoryService) processOrder(tx *gorm.DB, doc *models.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id is required for order")
	}
	for _, item := range doc.Items {
		available, err := s.GetAvailableQuantity(*doc.WarehouseID, item.VariantID)
		if err != nil {
			return err
		}

		if available.LessThan(item.Quantity) {
			return fmt.Errorf("not enough available stock for variant %d", item.VariantID)
		}

		res, err := s.reservationRepo.GetReservationWithTx(tx, *doc.WarehouseID, item.VariantID)
		if err != nil {
			return err
		}
		if res == nil {
			res = &models.StockReservation{WarehouseID: *doc.WarehouseID, VariantID: item.VariantID, Quantity: decimal.Zero}
		}
		res.Quantity = res.Quantity.Add(item.Quantity)
		if err := s.reservationRepo.SaveReservationWithTx(tx, res); err != nil {
			return err
		}
	}
	return nil
}

func (s *inventoryService) processReservationRelease(tx *gorm.DB, doc *models.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required")
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

func (s *inventoryService) revertOrder(tx *gorm.DB, doc *models.Document) error {
	return s.processReservationRelease(tx, doc)
}

func (s *inventoryService) revertReservationRelease(tx *gorm.DB, doc *models.Document) error {
	if doc.WarehouseID == nil {
		return errors.New("warehouse_id required")
	}
	for _, item := range doc.Items {
		res, err := s.reservationRepo.GetReservationWithTx(tx, *doc.WarehouseID, item.VariantID)
		if err != nil {
			return err
		}
		if res == nil {
			res = &models.StockReservation{WarehouseID: *doc.WarehouseID, VariantID: item.VariantID, Quantity: decimal.Zero}
		}
		res.Quantity = res.Quantity.Add(item.Quantity)
		if err := s.reservationRepo.SaveReservationWithTx(tx, res); err != nil {
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
