package models

import (
	"time"

	"github.com/shopspring/decimal"
)

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

type MovementFilter struct {
	VariantID *uint
	Limit     int
	Offset    int
}

type DocumentFilter struct {
	Search   *string
	Status   *string
	Types    []string
	DateFrom *time.Time
	DateTo   *time.Time

	Limit  int
	Offset int
}

type CounterpartyFilter struct {
	Search *string
	Limit  int
	Offset int
}
