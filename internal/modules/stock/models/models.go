package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type Product struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"unique;not null" json:"name"`
	Description string `json:"description"`
	CategoryID  uint   `json:"category_id"`
}

type CharacteristicsMap map[string]string

func (c CharacteristicsMap) Value() (driver.Value, error) { return json.Marshal(c) }
func (c *CharacteristicsMap) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

type Variant struct {
	ID              uint               `gorm:"primaryKey" json:"id"`
	ProductID       uint               `gorm:"index;not null" json:"product_id"`
	SKU             string             `gorm:"unique" json:"sku"`
	Characteristics CharacteristicsMap `gorm:"type:jsonb" json:"characteristics"`
	UnitID          uint               `json:"unit_id"`
}

type CharacteristicType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type CharacteristicValue struct {
	ID                   uint   `gorm:"primaryKey" json:"id"`
	CharacteristicTypeID uint   `gorm:"not null" json:"characteristic_type_id"`
	Value                string `gorm:"not null" json:"value"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Unit struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Warehouse struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"not null" json:"name"`
	Address string `json:"address"`
}

type Counterparty struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"unique;not null" json:"name"`
	Phone    string `gorm:"size:20" json:"phone"`
	Telegram string `gorm:"size:100" json:"telegram"`
	Email    string `gorm:"size:255" json:"email"`
	Address  string `json:"address"`
}

type PriceType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Document struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Type           string         `json:"type"`
	Number         string         `gorm:"uniqueIndex" json:"number"`
	WarehouseID    *uint          `json:"warehouse_id"`
	ToWarehouseID  *uint          `json:"to_warehouse_id"`
	CounterpartyID *uint          `json:"counterparty_id"`
	PriceTypeID    *uint          `json:"price_type_id"`
	Comment        string         `json:"comment"`
	BaseDocumentID *uint          `json:"base_document_id"`
	Items          []DocumentItem `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"items"`
	Status         string         `gorm:"default:draft" json:"status"`
	CreatedBy      *uint          `json:"created_by"`
	PostedAt       *time.Time     `json:"posted_at"`
	CreatedAt      time.Time      `json:"created_at"`
}

type DocumentItem struct {
	ID         uint             `gorm:"primaryKey" json:"id"`
	DocumentID uint             `json:"-"`
	VariantID  uint             `gorm:"column:item_id" json:"variant_id"`
	Quantity   decimal.Decimal  `gorm:"type:decimal(14,4);" json:"quantity"`
	Price      *decimal.Decimal `gorm:"type:decimal(14,2);" json:"price"`
}

type DocumentHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `json:"document_id"`
	Action     string    `json:"action"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  *uint     `json:"created_by"`
	Comment    string    `json:"comment"`
}

type StockMovement struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	DocumentID     *uint           `gorm:"index" json:"document_id"`
	VariantID      uint            `gorm:"column:item_id" json:"variant_id"`
	WarehouseID    uint            `json:"warehouse_id"`
	CounterpartyID *uint           `json:"counterparty_id"`
	Quantity       decimal.Decimal `gorm:"type:decimal(14,4);" json:"quantity"`
	// 	Cost           decimal.Decimal `gorm:"type:decimal(14,4);" json:"cost"`
	Type      string    `json:"type"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type StockBalance struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	WarehouseID uint            `gorm:"index:idx_wh_variant,unique" json:"warehouse_id"`
	VariantID   uint            `gorm:"column:item_id;index:idx_wh_variant,unique" json:"variant_id"`
	Quantity    decimal.Decimal `gorm:"type:decimal(14,4);" json:"quantity"`
	// TotalCost   decimal.Decimal `gorm:"type:decimal(18,4);" json:"total_cost"`
}

type StockReservation struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	WarehouseID uint            `gorm:"index:idx_wh_variant_res,unique" json:"warehouse_id"`
	VariantID   uint            `gorm:"index:idx_wh_variant_res,unique" json:"variant_id"`
	Quantity    decimal.Decimal `gorm:"type:decimal(14,4);" json:"quantity"`
}

type ItemPrice struct {
	VariantID   uint            `gorm:"column:item_id;primaryKey" json:"variant_id"`
	PriceTypeID uint            `gorm:"primaryKey" json:"price_type_id"`
	Price       decimal.Decimal `gorm:"type:decimal(14,2);" json:"price"`
	Currency    string          `gorm:"default:'RUB'" json:"currency"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type StockFilter struct {
	CategoryID *uint
	SKU        *string
	MinQty     *decimal.Decimal
}

type DocumentSequence struct {
	ID         string `gorm:"primaryKey"`
	LastNumber uint
}

// deprecated
type Item struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"unique;not null"`
	SKU        string `gorm:"unique"`
	UnitID     uint
	CategoryID uint
}

type StockLot struct {
	ID          uint `gorm:"primaryKey"`
	WarehouseID uint `gorm:"index:idx_wh_variant_lot"`
	VariantID   uint `gorm:"index:idx_wh_variant_lot"`

	IncomeDocumentID uint
	ArrivalDate      time.Time `gorm:"index"`

	CurrentQuantity decimal.Decimal `gorm:"type:decimal(14,4);"`
}
