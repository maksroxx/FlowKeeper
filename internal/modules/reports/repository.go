package reports

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Repository interface {
	GetFIFOProfitData(from, to time.Time, warehouseID *uint) ([]ProfitRecord, error)
	GetStockData(warehouseID *uint) ([]StockItem, error)
	GetMovementsData(from, to time.Time, warehouseID *uint) ([]MovementItem, error)
	GetCustomerData(from, to time.Time) ([]CustomerReportItem, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

type ProfitRecord struct {
	VariantID   uint
	SKU         string
	ProductName string
	Quantity    decimal.Decimal
	Revenue     decimal.Decimal
	Cost        decimal.Decimal
}

func (r *repository) GetFIFOProfitData(from, to time.Time, warehouseID *uint) ([]ProfitRecord, error) {
	var results []ProfitRecord
	query := r.db.Table("stock_movements as sm").
		Select(`
			variants.id as variant_id, variants.sku, products.name as product_name,
			COALESCE(SUM(ABS(sm.quantity)), 0) as quantity,
			COALESCE(SUM(ABS(sm.quantity) * COALESCE(sale_item.price, 0)), 0) as revenue,
			COALESCE(SUM(ABS(sm.quantity) * COALESCE(buy_item.price, 0)), 0) as cost
		`).
		Joins("JOIN variants ON variants.id = sm.item_id").
		Joins("JOIN products ON products.id = variants.product_id").
		Joins("LEFT JOIN document_items as sale_item ON sale_item.document_id = sm.document_id AND sale_item.item_id = sm.item_id").
		Joins("LEFT JOIN stock_lots as lot ON lot.id = sm.source_lot_id").
		Joins("LEFT JOIN document_items as buy_item ON buy_item.document_id = lot.income_document_id AND buy_item.item_id = lot.variant_id").
		Where("sm.type = ?", "OUTCOME").
		Where("sm.created_at BETWEEN ? AND ?", from, to)

	if warehouseID != nil {
		query = query.Where("sm.warehouse_id = ?", *warehouseID)
	}
	err := query.Group("variants.id, variants.sku, products.name").Scan(&results).Error
	return results, err
}

func (r *repository) GetStockData(warehouseID *uint) ([]StockItem, error) {
	var results []StockItem
	query := r.db.Table("stock_lots").
		Select(`
			warehouses.name as warehouse_name, categories.name as category,
			variants.sku, products.name as product_name, units.name as unit,
			SUM(stock_lots.current_quantity) as quantity,
			SUM(stock_lots.current_quantity * COALESCE(di.price, 0)) as total_value
		`).
		Joins("JOIN warehouses ON warehouses.id = stock_lots.warehouse_id").
		Joins("JOIN variants ON variants.id = stock_lots.variant_id").
		Joins("JOIN products ON products.id = variants.product_id").
		Joins("JOIN categories ON categories.id = products.category_id").
		Joins("JOIN units ON units.id = variants.unit_id").
		Joins("LEFT JOIN document_items di ON di.document_id = stock_lots.income_document_id AND di.item_id = stock_lots.variant_id").
		Where("stock_lots.current_quantity > 0")

	if warehouseID != nil {
		query = query.Where("stock_lots.warehouse_id = ?", *warehouseID)
	}
	err := query.Group("warehouses.name, categories.name, variants.sku, products.name, units.name").
		Order("warehouses.name, categories.name, products.name").
		Scan(&results).Error
	return results, err
}

func (r *repository) GetMovementsData(from, to time.Time, warehouseID *uint) ([]MovementItem, error) {
	var rows []struct {
		Date      time.Time
		DocType   string
		DocNumber string
		WhName    string
		Sku       string
		ProdName  string
		UnitName  string
		Qty       decimal.Decimal
		Type      string
	}
	query := r.db.Table("stock_movements").
		Select(`
			stock_movements.created_at as date, documents.type as doc_type, documents.number as doc_number,
			warehouses.name as wh_name, variants.sku as sku, products.name as prod_name,
			units.name as unit_name, stock_movements.quantity as qty, stock_movements.type as type
		`).
		Joins("LEFT JOIN documents ON documents.id = stock_movements.document_id").
		Joins("JOIN warehouses ON warehouses.id = stock_movements.warehouse_id").
		Joins("JOIN variants ON variants.id = stock_movements.item_id").
		Joins("JOIN products ON products.id = variants.product_id").
		Joins("JOIN units ON units.id = variants.unit_id").
		Where("stock_movements.created_at BETWEEN ? AND ?", from, to)

	if warehouseID != nil {
		query = query.Where("stock_movements.warehouse_id = ?", *warehouseID)
	}
	if err := query.Order("stock_movements.created_at desc").Scan(&rows).Error; err != nil {
		return nil, err
	}

	var results []MovementItem
	for _, row := range rows {
		item := MovementItem{
			Date: row.Date, DocumentType: row.DocType, DocumentNumber: row.DocNumber,
			WarehouseName: row.WhName, SKU: row.Sku, ProductName: row.ProdName, Unit: row.UnitName,
		}
		if row.Type == "INCOME" || (row.Type == "TRANSFER" && row.Qty.IsPositive()) {
			item.QuantityIn = row.Qty.Abs()
		} else {
			item.QuantityOut = row.Qty.Abs()
		}
		results = append(results, item)
	}
	return results, nil
}

func (r *repository) GetCustomerData(from, to time.Time) ([]CustomerReportItem, error) {
	var results []CustomerReportItem
	err := r.db.Table("documents").
		Select(`
            COALESCE(counterparties.name, 'Розничный покупатель') as counterparty_name,
            COUNT(documents.id) as operations_count,
            COALESCE(SUM(di.quantity * di.price), 0) as total_revenue
        `).
		Joins("LEFT JOIN counterparties ON counterparties.id = documents.counterparty_id").
		Joins("JOIN document_items di ON di.document_id = documents.id").
		Where("documents.type = ? AND documents.status = ?", "OUTCOME", "posted").
		Where("documents.created_at BETWEEN ? AND ?", from, to).
		Group("COALESCE(counterparties.name, 'Розничный покупатель')").
		Order("total_revenue DESC").
		Scan(&results).Error

	return results, err
}
