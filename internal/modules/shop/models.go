package shop

import "github.com/maksroxx/flowkeeper/internal/modules/stock/models"

// Настройки магазина
type ShopSettings struct {
	ID                  uint   `gorm:"primaryKey" json:"id"`
	IsEnabled           bool   `json:"is_enabled"`
	StockDisplayRule    string `json:"stock_display_rule" gorm:"default:'all'"` // all, in_stock, out_of_stock
	ShowPrice           bool   `json:"show_price"`
	ShowStock           bool   `json:"show_stock"`
	ShowSKU             bool   `json:"show_sku"`
	ShowDesc            bool   `json:"show_desc"`
	ShowCharacteristics bool   `json:"show_characteristics"`
}

// DTO для одной вариации (Внутренний объект массива variants)
type ShopVariantDTO struct {
	ID              uint              `json:"id"`
	SKU             string            `json:"sku"`
	Price           float64           `json:"price"`
	Currency        string            `json:"currency"`
	Quantity        float64           `json:"quantity"`
	Images          []string          `json:"images"`          // Картинки конкретно этой вариации
	Characteristics map[string]string `json:"characteristics"` // Цвет: Черный и т.д.
}

// DTO для товара (Главный объект)
type ShopProductDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsFeatured  bool   `json:"is_featured"`

	// Общий список картинок (для слайдера в каталоге или общей галереи)
	AllImages []string `json:"images"`

	// Поля для каталога (краткая сводка)
	PriceFrom float64 `json:"price,omitempty"`    // Минимальная цена
	Currency  string  `json:"currency,omitempty"` // Валюта
	TotalQty  float64 `json:"quantity,omitempty"` // Общий остаток

	// Поле для детальной карточки (список вариаций)
	Variants []ShopVariantDTO `json:"variants,omitempty"`
}

// Вспомогательная функция для конвертации
func ToVariantDTO(v models.Variant, price float64, currency string, qty float64) ShopVariantDTO {
	var images []string
	if len(v.Images) > 0 {
		for _, img := range v.Images {
			images = append(images, img.URL)
		}
	} else {
		images = make([]string, 0)
	}

	if currency == "" {
		currency = "RUB"
	}

	return ShopVariantDTO{
		ID:              v.ID,
		SKU:             v.SKU,
		Price:           price,
		Currency:        currency,
		Quantity:        qty,
		Images:          images,
		Characteristics: v.Characteristics,
	}
}
