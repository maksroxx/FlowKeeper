package shop

import (
	"errors"
	"strings"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Repository interface {
	GetSettings() (*ShopSettings, error)
	UpdateSettings(s *ShopSettings) error
	GetCatalog(search string) ([]ShopProductDTO, error)
	GetProductDetails(id uint) (*ShopProductDTO, error)
	ToggleProductStatus(productID uint, isPublic, isFeatured *bool) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetSettings() (*ShopSettings, error) {
	var s ShopSettings
	err := r.db.First(&s, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s = ShopSettings{ID: 1, ShowPrice: true, ShowStock: true, ShowSKU: true, ShowDesc: true}
		r.db.Create(&s)
	}
	return &s, nil
}

func (r *repository) UpdateSettings(s *ShopSettings) error {
	s.ID = 1
	return r.db.Save(s).Error
}

func (r *repository) GetCatalog(search string) ([]ShopProductDTO, error) {
	settings, err := r.GetSettings()
	if err != nil {
		return nil, err
	}

	var products []models.Product

	query := r.db.Model(&models.Product{}).
		Preload("Variants").
		Preload("Variants.Images").
		Where("is_public = ?", true)

	if search != "" {
		term := "%" + strings.ToLower(search) + "%"
		query = query.Joins("LEFT JOIN variants ON variants.product_id = products.id").
			Where("LOWER(products.name) LIKE ? OR LOWER(variants.sku) LIKE ?", term, term).
			Group("products.id")
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	var results []ShopProductDTO

	for _, p := range products {
		if len(p.Variants) == 0 {
			continue
		}

		var allImages []string
		var minPrice float64 = -1
		var totalQty float64 = 0
		var mainCurrency string = "RUB"
		hasStock := false

		for _, v := range p.Variants {
			for _, img := range v.Images {
				allImages = append(allImages, img.URL)
			}

			price, currency, qty := r.getVariantData(v.ID)

			if mainCurrency == "RUB" && currency != "" {
				mainCurrency = currency
			}

			if minPrice == -1 || price < minPrice {
				minPrice = price
			}
			totalQty += qty

			if qty > 0 {
				hasStock = true
			}
		}

		if settings.StockDisplayRule == "in_stock" && !hasStock {
			continue
		}
		if settings.StockDisplayRule == "out_of_stock" && hasStock {
			continue
		}

		if allImages == nil {
			allImages = make([]string, 0)
		}

		results = append(results, ShopProductDTO{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			IsFeatured:  p.IsFeatured,
			AllImages:   allImages,
			PriceFrom:   minPrice,
			Currency:    mainCurrency,
			TotalQty:    totalQty,
		})
	}

	return results, nil
}

func (r *repository) GetProductDetails(id uint) (*ShopProductDTO, error) {
	settings, err := r.GetSettings()
	if err != nil {
		return nil, err
	}

	var p models.Product
	err = r.db.Model(&models.Product{}).
		Where("id = ? AND is_public = ?", id, true).
		Preload("Variants").
		Preload("Variants.Images").
		First(&p).Error

	if err != nil {
		return nil, err
	}

	var variantsDTO []ShopVariantDTO
	var allImages []string
	var totalQty float64
	var minPrice float64 = -1
	var currency string = "RUB"
	var hasStock bool = false

	for _, v := range p.Variants {
		price, curr, qty := r.getVariantData(v.ID)
		if curr != "" {
			currency = curr
		}

		totalQty += qty
		if qty > 0 {
			hasStock = true
		}

		if minPrice == -1 || price < minPrice {
			minPrice = price
		}

		vDto := ToVariantDTO(v, price, currency, qty)
		variantsDTO = append(variantsDTO, vDto)

		allImages = append(allImages, vDto.Images...)
	}

	if settings.StockDisplayRule == "in_stock" && !hasStock {
		return nil, errors.New("product out of stock")
	}

	if allImages == nil {
		allImages = make([]string, 0)
	}

	dto := &ShopProductDTO{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		IsFeatured:  p.IsFeatured,
		AllImages:   allImages,
		PriceFrom:   minPrice,
		Currency:    currency,
		TotalQty:    totalQty,
		Variants:    variantsDTO,
	}

	return dto, nil
}

func (r *repository) getVariantData(variantID uint) (float64, string, float64) {
	var priceModel models.ItemPrice
	r.db.Where("item_id = ?", variantID).First(&priceModel)

	var bal decimal.Decimal
	r.db.Model(&models.StockBalance{}).
		Where("item_id = ?", variantID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&bal)

	return priceModel.Price.InexactFloat64(), priceModel.Currency, bal.InexactFloat64()
}

func (r *repository) ToggleProductStatus(productID uint, isPublic, isFeatured *bool) error {
	updates := make(map[string]interface{})
	if isPublic != nil {
		updates["is_public"] = *isPublic
	}
	if isFeatured != nil {
		updates["is_featured"] = *isFeatured
	}
	return r.db.Model(&models.Product{}).Where("id = ?", productID).Updates(updates).Error
}
