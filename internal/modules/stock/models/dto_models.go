package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type VariantDTO struct {
	ID              uint               `json:"id"`
	ProductID       uint               `json:"product_id"`
	ProductName     string             `json:"product_name"`
	SKU             string             `json:"sku"`
	Characteristics CharacteristicsMap `json:"characteristics"`
	UnitID          uint               `json:"unit_id"`
	UnitName        string             `json:"unit_name"`
}

type VariantListItemDTO struct {
	ID              uint               `json:"id"`
	ProductID       uint               `json:"product_id"`
	ProductName     string             `json:"product_name"`
	SKU             string             `json:"sku"`
	Characteristics CharacteristicsMap `json:"characteristics"`
	UnitID          uint               `json:"unit_id"`
	UnitName        string             `json:"unit_name"`

	QuantityOnStock decimal.Decimal `json:"quantity_on_stock"`
}

type DocumentDTO struct {
	ID               uint              `json:"id"`
	Type             string            `json:"type"`
	Number           string            `json:"number"`
	WarehouseID      *uint             `json:"warehouse_id,omitempty"`
	WarehouseName    string            `json:"warehouse_name,omitempty"`
	ToWarehouseID    *uint             `json:"to_warehouse_id,omitempty"`
	ToWarehouseName  string            `json:"to_warehouse_name,omitempty"`
	CounterpartyID   *uint             `json:"counterparty_id,omitempty"`
	CounterpartyName string            `json:"counterparty_name,omitempty"`
	PriceTypeID      *uint             `json:"price_type_id,omitempty"`
	PriceTypeName    string            `json:"price_type_name,omitempty"`
	Comment          string            `json:"comment"`
	BaseDocumentID   *uint             `json:"base_document_id,omitempty"`
	Items            []DocumentItemDTO `json:"items"`
	Status           string            `json:"status"`
	PostedAt         *time.Time        `json:"posted_at,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}

type DocumentItemDTO struct {
	ID          uint             `json:"id"`
	VariantID   uint             `json:"variant_id"`
	VariantSKU  string           `json:"variant_sku"`
	ProductName string           `json:"product_name"`
	Quantity    decimal.Decimal  `json:"quantity"`
	Price       *decimal.Decimal `json:"price,omitempty"`
}

type DocumentListItemDTO struct {
	ID               uint      `json:"id"`
	Type             string    `json:"type"`
	Number           string    `json:"number"`
	WarehouseName    string    `json:"warehouse_name,omitempty"`
	CounterpartyName string    `json:"counterparty_name,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type StockMovementDTO struct {
	ID             uint            `json:"id"`
	DocumentID     *uint           `json:"document_id"`
	DocumentNumber string          `json:"document_number,omitempty"`
	VariantID      uint            `json:"variant_id"`
	VariantSKU     string          `json:"variant_sku"`
	ProductName    string          `json:"product_name"`
	WarehouseID    uint            `json:"warehouse_id"`
	WarehouseName  string          `json:"warehouse_name"`
	Quantity       decimal.Decimal `json:"quantity"`
	Type           string          `json:"type"`
	CreatedAt      time.Time       `json:"created_at"`
}

type StockBalanceDTO struct {
	ID          uint            `json:"id"`
	WarehouseID uint            `json:"warehouse_id"`
	VariantID   uint            `json:"variant_id"`
	VariantSKU  string          `json:"variant_sku"`
	ProductName string          `json:"product_name"`
	UnitName    string          `json:"unit_name"`
	Quantity    decimal.Decimal `json:"quantity"`
}
