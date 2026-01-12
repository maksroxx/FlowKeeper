package analytics

import (
	"time"

	"github.com/shopspring/decimal"
)

type DashboardData struct {
	TotalStock decimal.Decimal `json:"total_stock"`

	TotalItemsCount  int64 `json:"total_items_count"`
	RecentOperations int64 `json:"recent_operations"`

	// Метрики "здоровья" склада
	TotalVariants int64 `json:"total_variants"`
	ItemsInStock  int64 `json:"items_in_stock"`
	LowStockCount int64 `json:"low_stock_count"`

	IncomingToday decimal.Decimal `json:"incoming_today"`
	OutgoingToday decimal.Decimal `json:"outgoing_today"`

	ChartData       []ChartPoint    `json:"chart_data"`
	RecentMovements []MovementShort `json:"recent_movements"`
}

type ChartPoint struct {
	Date  string          `json:"date"`
	Value decimal.Decimal `json:"value"`
	Type  string          `json:"type"` // INCOME / OUTCOME
}

type MovementShort struct {
	ID            uint            `json:"id"`
	Date          time.Time       `json:"date"`
	ItemName      string          `json:"item_name"`
	Type          string          `json:"type"`
	Quantity      decimal.Decimal `json:"quantity"`
	WarehouseName string          `json:"warehouse_name"`
}
