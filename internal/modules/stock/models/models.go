package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Item struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"unique;not null"`
	SKU        string `gorm:"unique"`
	UnitID     uint
	CategoryID uint
}

type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type Unit struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type Warehouse struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"not null"`
	Address string
}

type Counterparty struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"unique;not null"`
	Phone    string `gorm:"size:20"`
	Telegram string `gorm:"size:100"`
	Email    string `gorm:"size:255"`
	Address  string
}

type Document struct {
	ID             uint   `gorm:"primaryKey"`
	Type           string // INCOME, OUTCOME, TRANSFER, INVENTORY, PRICE_UPDATE
	Number         string `gorm:"uniqueIndex"`
	WarehouseID    *uint
	ToWarehouseID  *uint
	CounterpartyID *uint
	PriceTypeID    *uint
	Comment        string
	Items          []DocumentItem `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE"`
	Status         string         `gorm:"default:draft"` // draft|posted|canceled
	CreatedBy      *uint
	PostedAt       *time.Time
	CreatedAt      time.Time
}

type DocumentItem struct {
	ID         uint `gorm:"primaryKey"`
	DocumentID uint
	ItemID     uint
	Quantity   decimal.Decimal  `gorm:"type:decimal(14,4);"`
	Price      *decimal.Decimal `gorm:"type:decimal(14,2);"`
}

type DocumentHistory struct {
	ID         uint `gorm:"primaryKey"`
	DocumentID uint
	Action     string // "posted", "canceled"
	CreatedAt  time.Time
	CreatedBy  *uint
	Comment    string
}

type StockMovement struct {
	ID             uint  `gorm:"primaryKey"`
	DocumentID     *uint `gorm:"index"`
	ItemID         uint
	WarehouseID    uint
	CounterpartyID *uint
	Quantity       decimal.Decimal `gorm:"type:decimal(14,4);"`
	Cost           decimal.Decimal `gorm:"type:decimal(14,4);"`
	Type           string          // income/outcome/transfer/inventory/cancel
	Comment        string
	CreatedAt      time.Time
}

type StockBalance struct {
	ID          uint            `gorm:"primaryKey"`
	WarehouseID uint            `gorm:"index:idx_wh_item,unique"`
	ItemID      uint            `gorm:"index:idx_wh_item,unique"`
	Quantity    decimal.Decimal `gorm:"type:decimal(14,4);"`
	TotalCost   decimal.Decimal `gorm:"type:decimal(18,4);"`
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

type PriceType struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type ItemPrice struct {
	ItemID      uint `gorm:"primaryKey"`
	PriceTypeID uint `gorm:"primaryKey"`

	Price     decimal.Decimal `gorm:"type:decimal(14,2);"`
	Currency  string          `gorm:"default:'RUB'"`
	UpdatedAt time.Time
}
