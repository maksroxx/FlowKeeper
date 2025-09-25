package stock

import "time"

type Item struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"unique;not null"`
	Stock int
}

type StockMovement struct {
	ID        uint `gorm:"primaryKey"`
	ItemID    uint
	Quantity  int
	CreatedAt time.Time
}
