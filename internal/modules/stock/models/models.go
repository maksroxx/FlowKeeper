package models

import "time"

type Item struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"unique;not null"`
	SKU        string `gorm:"unique"`
	UnitID     uint
	CategoryID uint
	Price      float64
}

type DocumentHistory struct {
	ID         uint `gorm:"primaryKey"`
	DocumentID uint
	Action     string // "posted", "canceled"
	CreatedAt  time.Time
	CreatedBy  *uint
	Comment    string
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
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

// Document + items
type Document struct {
	ID             uint   `gorm:"primaryKey"`
	Type           string // INCOME, OUTCOME, TRANSFER, INVENTORY
	Number         string `gorm:"uniqueIndex"`
	WarehouseID    *uint
	ToWarehouseID  *uint
	CounterpartyID *uint
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
	Quantity   int
}

// StockMovement: теперь с DocumentID
type StockMovement struct {
	ID             uint  `gorm:"primaryKey"`
	DocumentID     *uint `gorm:"index"`
	ItemID         uint
	WarehouseID    uint
	CounterpartyID *uint
	Quantity       int    // signed: + for in, - for out
	Type           string // income/outcome/transfer/inventory/cancel
	Comment        string
	CreatedAt      time.Time
}

// StockBalance
type StockBalance struct {
	ID          uint `gorm:"primaryKey"`
	WarehouseID uint `gorm:"index:idx_wh_item,unique"`
	ItemID      uint `gorm:"index:idx_wh_item,unique"`
	Quantity    int
}

type StockFilter struct {
	CategoryID *uint
	SKU        *string
	MinQty     *int
}

type DocumentSequence struct {
	// ID будет составным, например "INCOME_2025"
	ID         string `gorm:"primaryKey"`
	LastNumber uint
}
