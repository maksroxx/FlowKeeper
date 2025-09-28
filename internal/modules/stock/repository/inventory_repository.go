package repository

import (
	"gorm.io/gorm"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type BalanceRepository interface {
	GetBalance(warehouseID, itemID uint) (*stock.StockBalance, error)
	GetBalanceWithTx(tx *gorm.DB, warehouseID, itemID uint) (*stock.StockBalance, error)

	SaveBalance(b *stock.StockBalance) error
	SaveBalanceWithTx(tx *gorm.DB, b *stock.StockBalance) error

	ListByWarehouse(warehouseID uint) ([]stock.StockBalance, error)
	ListByWarehouseFiltered(warehouseID uint, f stock.StockFilter) ([]stock.StockBalance, error)
}

type balanceRepo struct {
	db *gorm.DB
}

func NewBalanceRepository(db *gorm.DB) BalanceRepository { return &balanceRepo{db: db} }

func (r *balanceRepo) GetBalance(warehouseID, itemID uint) (*stock.StockBalance, error) {
	return r.GetBalanceWithTx(nil, warehouseID, itemID)
}

func (r *balanceRepo) GetBalanceWithTx(tx *gorm.DB, warehouseID, itemID uint) (*stock.StockBalance, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var b stock.StockBalance
	if err := db.Where("warehouse_id = ? AND item_id = ?", warehouseID, itemID).First(&b).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *balanceRepo) SaveBalance(b *stock.StockBalance) error {
	return r.SaveBalanceWithTx(nil, b)
}

func (r *balanceRepo) SaveBalanceWithTx(tx *gorm.DB, b *stock.StockBalance) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(b).Error
}

func (r *balanceRepo) ListByWarehouse(warehouseID uint) ([]stock.StockBalance, error) {
	var bs []stock.StockBalance
	if err := r.db.Where("warehouse_id = ?", warehouseID).Find(&bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

func (r *balanceRepo) ListByWarehouseFiltered(warehouseID uint, f stock.StockFilter) ([]stock.StockBalance, error) {
	var balances []stock.StockBalance

	query := r.db.Model(&stock.StockBalance{}).
		Where("warehouse_id = ?", warehouseID)

	if f.MinQty != nil {
		query = query.Where("quantity >= ?", *f.MinQty)
	}

	if f.SKU != nil || f.CategoryID != nil {
		query = query.Joins("JOIN items ON items.id = stock_balances.item_id")

		if f.SKU != nil {
			query = query.Where("items.sku = ?", *f.SKU)
		}
		if f.CategoryID != nil {
			query = query.Where("items.category_id = ?", *f.CategoryID)
		}
	}

	if err := query.Find(&balances).Error; err != nil {
		return nil, err
	}

	return balances, nil
}
