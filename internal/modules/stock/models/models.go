package stock

import "time"

// Item
type Item struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"unique;not null"`
	SKU        string `gorm:"unique"`
	UnitID     uint
	CategoryID uint
	Price      float64
}

// Category
type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

// Unit
type Unit struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

// Warehouse
type Warehouse struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"not null"`
	Address string
}

// Counterparty
type Counterparty struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

// StockMovement
type StockMovement struct {
	ID             uint `gorm:"primaryKey"`
	ItemID         uint
	WarehouseID    uint
	CounterpartyID *uint
	Quantity       int
	Type           string
	Comment        string
	CreatedAt      time.Time
}

// Document
type Document struct {
	ID             uint `gorm:"primaryKey"`
	Type           string
	Number         string
	WarehouseID    *uint
	CounterpartyID *uint
	Comment        string
	Items          []DocumentItem `gorm:"foreignKey:DocumentID"`
	CreatedAt      time.Time
}

// DocumentItem
type DocumentItem struct {
	ID         uint `gorm:"primaryKey"`
	DocumentID uint
	ItemID     uint
	Quantity   int
}
