package analytics

import (
	"time"

	stockModels "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Repository interface {
	GetTotalStock(warehouseID *uint) (decimal.Decimal, int64, error)
	GetActivityStats(warehouseID *uint, days int) (int64, decimal.Decimal, decimal.Decimal, error)
	GetChartData(warehouseID *uint, days int) ([]stockModels.StockMovement, error)
	GetInventoryHealth(warehouseID *uint) (int64, int64, int64, error)
	GetRecentMovements(warehouseID *uint, limit int) ([]MovementShort, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetTotalStock(warehouseID *uint) (decimal.Decimal, int64, error) {
	var balances []stockModels.StockBalance
	query := r.db.Model(&stockModels.StockBalance{})

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if err := query.Find(&balances).Error; err != nil {
		return decimal.Zero, 0, err
	}

	totalQty := decimal.Zero
	for _, b := range balances {
		totalQty = totalQty.Add(b.Quantity)
	}

	return totalQty, int64(len(balances)), nil
}

func (r *repository) GetActivityStats(warehouseID *uint, days int) (int64, decimal.Decimal, decimal.Decimal, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	startToday := time.Now().Truncate(24 * time.Hour)

	var movements []stockModels.StockMovement
	query := r.db.Model(&stockModels.StockMovement{}).Where("created_at >= ?", startDate)

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if err := query.Find(&movements).Error; err != nil {
		return 0, decimal.Zero, decimal.Zero, err
	}

	recentOps := int64(len(movements))
	inToday := decimal.Zero
	outToday := decimal.Zero

	for _, m := range movements {
		if m.CreatedAt.After(startToday) {
			if m.Type == "INCOME" {
				inToday = inToday.Add(m.Quantity)
			} else if m.Type == "OUTCOME" {
				outToday = outToday.Add(m.Quantity.Abs())
			}
		}
	}

	return recentOps, inToday, outToday, nil
}

func (r *repository) GetChartData(warehouseID *uint, days int) ([]stockModels.StockMovement, error) {
	var movements []stockModels.StockMovement
	startDate := time.Now().AddDate(0, 0, -days)

	query := r.db.Model(&stockModels.StockMovement{}).Where("created_at >= ?", startDate)
	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	err := query.Order("created_at asc").Find(&movements).Error
	return movements, err
}

func (r *repository) GetInventoryHealth(warehouseID *uint) (int64, int64, int64, error) {
	var total int64
	var inStock int64
	var lowStock int64

	r.db.Model(&stockModels.Variant{}).Count(&total)

	query := r.db.Model(&stockModels.StockBalance{})
	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	query.Where("quantity > 0").Count(&inStock)
	queryLow := r.db.Model(&stockModels.StockBalance{}).Where("quantity > 0 AND quantity < 10")
	if warehouseID != nil {
		queryLow = queryLow.Where("warehouse_id = ?", *warehouseID)
	}
	queryLow.Count(&lowStock)

	return total, inStock, lowStock, nil
}

func (r *repository) GetRecentMovements(warehouseID *uint, limit int) ([]MovementShort, error) {
	var results []MovementShort

	query := r.db.Table("stock_movements").
		Select("stock_movements.id, stock_movements.created_at as date, stock_movements.type, stock_movements.quantity, products.name as item_name, warehouses.name as warehouse_name").
		Joins("LEFT JOIN variants ON stock_movements.item_id = variants.id").
		Joins("LEFT JOIN products ON variants.product_id = products.id").
		Joins("LEFT JOIN warehouses ON stock_movements.warehouse_id = warehouses.id")

	if warehouseID != nil {
		query = query.Where("stock_movements.warehouse_id = ?", *warehouseID)
	}

	err := query.Order("stock_movements.created_at desc").
		Limit(limit).
		Scan(&results).Error

	return results, err
}
