package models

import "github.com/shopspring/decimal"

type VariantFilter struct {
	Name       *string
	CategoryID *uint

	SKU *string
	// "all" (по умолчанию), "in_stock", "out_of_stock"
	StockStatus string
	WarehouseID *uint

	Limit  int
	Offset int
}

type StockFilter struct {
	CategoryID *uint
	SKU        *string
	MinQty     *decimal.Decimal
}
