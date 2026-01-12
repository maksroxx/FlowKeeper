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
	variantRepo  repository.VariantRepository
	productRepo  repository.ProductRepository
}

func NewStrategyFactory(
	b repository.BalanceRepository,
	m repository.StockMovementRepository,
	l repository.LotRepository,
	v repository.VariantRepository,
	p repository.ProductRepository,
) StrategyFactory {
	return &strategyFactoryImpl{
		balanceRepo:  b,
		movementRepo: m,
		lotRepo:      l,
		variantRepo:  v,
		productRepo:  p,
	}
}

func (f *strategyFactoryImpl) GetStrategy(policy string) (QuantityStrategy, error) {
	switch policy {
	case "total":
		return NewTotalQuantityStrategy(f.balanceRepo, f.movementRepo), nil
	case "fifo":
		return NewFifoQuantityStrategy(f.lotRepo, f.movementRepo, f.balanceRepo, f.variantRepo, f.productRepo), nil
	default:
		return nil, fmt.Errorf("unknown quantity accounting policy: %s", policy)
	}
}

type InventoryService interface {
	ProcessDocumentWithTx(tx *gorm.DB, doc *models.Document) error
	RevertDocumentWithTx(tx *gorm.DB, doc *models.Document) error
	GetAvailableQuantity(warehouseID, variantID uint) (decimal.Decimal, error)
	ListByWarehouseFilteredAsDTO(warehouseID uint, f models.StockFilter) ([]models.StockBalanceDTO, error)

	GetStockByVariant(variantID uint) ([]models.VariantStockDTO, error)
}
type inventoryService struct {
	strategyFactory StrategyFactory
	reservationRepo repository.ReservationRepository
	balanceRepo     repository.BalanceRepository
	config          *config.Config
	whRepo          repository.WarehouseRepository

	variantRepo repository.VariantRepository
	productRepo repository.ProductRepository
	unitRepo    repository.UnitRepository
	catRepo     repository.CategoryRepository
}

func NewInventoryService(
	factory StrategyFactory,
	r repository.ReservationRepository,
	b repository.BalanceRepository,
	cfg *config.Config,
	v repository.VariantRepository,
	p repository.ProductRepository,
	u repository.UnitRepository,
	catRepo repository.CategoryRepository,
	whRepo repository.WarehouseRepository,
) InventoryService {
	return &inventoryService{
		strategyFactory: factory,
		reservationRepo: r,
		balanceRepo:     b,
		config:          cfg,
		variantRepo:     v,
		productRepo:     p,
		unitRepo:        u,
		catRepo:         catRepo,
		whRepo:          whRepo,
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

	variantIDsMap := make(map[uint]bool)
	for _, b := range balances {
		variantIDsMap[b.VariantID] = true
	}
	variantIDs := mapKeysToSlice(variantIDsMap)

	variants, _ := s.variantRepo.GetByIDs(variantIDs)
	variantMap := make(map[uint]models.Variant, len(variants))
	productIDsMap := make(map[uint]bool)
	unitIDsMap := make(map[uint]bool)
	for _, v := range variants {
		variantMap[v.ID] = v
		productIDsMap[v.ProductID] = true
		unitIDsMap[v.UnitID] = true
	}

	products, _ := s.productRepo.GetByIDs(mapKeysToSlice(productIDsMap))
	productMap := make(map[uint]models.Product, len(products))
	categoryIDsMap := make(map[uint]bool)
	for _, p := range products {
		productMap[p.ID] = p
		categoryIDsMap[p.CategoryID] = true
	}

	categories, _ := s.catRepo.GetByIDs(mapKeysToSlice(categoryIDsMap))
	categoryMap := make(map[uint]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
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
			ID:           b.ID,
			WarehouseID:  b.WarehouseID,
			VariantID:    b.VariantID,
			VariantSKU:   variant.SKU,
			ProductName:  product.Name,
			CategoryID:   product.CategoryID,
			CategoryName: categoryMap[product.CategoryID],
			UnitName:     unitMap[variant.UnitID],
			Quantity:     b.Quantity,
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
			variant, _ := s.variantRepo.GetByID(item.VariantID)
			productName := "Unknown Item"
			sku := "No SKU"
			if variant != nil {
				sku = variant.SKU
				if p, _ := s.productRepo.GetByID(variant.ProductID); p != nil {
					productName = p.Name
				}
			}
			return fmt.Errorf("Недостаточно товара '%s' (%s) на складе. Доступно: %s, Требуется: %s",
				productName, sku, available.String(), item.Quantity.String())
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

func (s *inventoryService) GetStockByVariant(variantID uint) ([]models.VariantStockDTO, error) {
	warehouses, err := s.whRepo.List()
	if err != nil {
		return nil, fmt.Errorf("could not list warehouses: %w", err)
	}

	if len(warehouses) == 0 {
		return []models.VariantStockDTO{}, nil
	}

	results := make([]models.VariantStockDTO, len(warehouses))

	for i, wh := range warehouses {
		balance, err := s.balanceRepo.GetBalanceWithTx(nil, wh.ID, variantID)
		if err != nil {
			return nil, err
		}

		reservation, err := s.reservationRepo.GetReservationWithTx(nil, wh.ID, variantID)
		if err != nil {
			return nil, err
		}

		onHandQty := decimal.Zero
		if balance != nil {
			onHandQty = balance.Quantity
		}

		reservedQty := decimal.Zero
		if reservation != nil {
			reservedQty = reservation.Quantity
		}

		results[i] = models.VariantStockDTO{
			WarehouseID:   wh.ID,
			WarehouseName: wh.Name,
			OnHand:        onHandQty,
			Reserved:      reservedQty,
			Available:     onHandQty.Sub(reservedQty),
		}
	}

	return results, nil
}

func toUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s)
}
