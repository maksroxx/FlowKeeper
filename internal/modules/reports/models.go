package reports

import (
	"time"

	"github.com/shopspring/decimal"
)

// ReportRequest - Параметры запроса от фронтенда
type ReportRequest struct {
	Type        string          `form:"type"`         // Тип отчета: "profit"
	DateFrom    time.Time       `form:"date_from"`    // Дата начала
	DateTo      time.Time       `form:"date_to"`      // Дата конца
	WarehouseID *uint           `form:"warehouse_id"` // ID склада (опционально)
	Format      string          `form:"format"`       // "pdf"
	TaxRate     decimal.Decimal `form:"tax_rate"`     // Ставка налога (например, 15.0)
}

// ProfitItem - Строка отчета о прибыли
type ProfitItem struct {
	SKU           string          `json:"sku"`
	ProductName   string          `json:"product_name"`
	QuantitySold  decimal.Decimal `json:"quantity_sold"`
	SalesTotal    decimal.Decimal `json:"sales_total"`   // Выручка
	CostTotal     decimal.Decimal `json:"cost_total"`    // Себестоимость (FIFO)
	GrossProfit   decimal.Decimal `json:"gross_profit"`  // Валовая прибыль
	EstimatedTax  decimal.Decimal `json:"estimated_tax"` // Налог
	NetProfit     decimal.Decimal `json:"net_profit"`    // Чистая прибыль
	Profitability decimal.Decimal `json:"profitability"` // Рентабельность %
}

// Остатки
type StockItem struct {
	WarehouseName string          `json:"warehouse_name"`
	Category      string          `json:"category"`
	SKU           string          `json:"sku"`
	ProductName   string          `json:"product_name"`
	Unit          string          `json:"unit"`
	Quantity      decimal.Decimal `json:"quantity"`
	//
	AvgPurchasePrice decimal.Decimal `json:"avg_purchase_price"`
	TotalValue       decimal.Decimal `json:"total_value"`
}

// Движения
type MovementItem struct {
	Date           time.Time       `json:"date"`
	DocumentType   string          `json:"document_type"` // INCOME, OUTCOME, ORDER...
	DocumentNumber string          `json:"document_number"`
	WarehouseName  string          `json:"warehouse_name"`
	SKU            string          `json:"sku"`
	ProductName    string          `json:"product_name"`
	QuantityIn     decimal.Decimal `json:"quantity_in"`  // Приход
	QuantityOut    decimal.Decimal `json:"quantity_out"` // Расход
	Unit           string          `json:"unit"`
}

// Продажи по клиентам
type CustomerReportItem struct {
	CounterpartyName string          `json:"counterparty_name"`
	OperationsCount  int64           `json:"operations_count"`
	TotalRevenue     decimal.Decimal `json:"total_revenue"`
}

// ABC
type ABCItem struct {
	SKU          string          `json:"sku"`
	ProductName  string          `json:"product_name"`
	QuantitySold decimal.Decimal `json:"quantity_sold"`
	Revenue      decimal.Decimal `json:"revenue"`       // Выручка
	SharePercent decimal.Decimal `json:"share_percent"` // Доля в общей выручке %
	Class        string          `json:"class"`         // A, B, C
}
