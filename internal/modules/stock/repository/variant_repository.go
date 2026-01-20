package repository

import (
	"strings"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type VariantRepository interface {
	Create(v *models.Variant) (*models.Variant, error)
	GetByID(id uint) (*models.Variant, error)
	GetByIDs(ids []uint) ([]models.Variant, error)
	List() ([]models.Variant, error)
	Update(v *models.Variant) (*models.Variant, error)
	Patch(id uint, updates map[string]interface{}) (*models.Variant, error)
	Delete(id uint) error
	Search(filter models.VariantFilter) ([]models.VariantListItemDTO, error)
	FindByProductID(productID uint) ([]models.Variant, error)
}

type variantRepo struct{ db *gorm.DB }

func NewVariantRepository(db *gorm.DB) VariantRepository {
	return &variantRepo{db: db}
}

func (r *variantRepo) Create(v *models.Variant) (*models.Variant, error) {
	err := r.db.Create(v).Error
	return v, err
}

func (r *variantRepo) GetByID(id uint) (*models.Variant, error) {
	var variant models.Variant
	if err := r.db.Preload("Images").First(&variant, id).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *variantRepo) GetByIDs(ids []uint) ([]models.Variant, error) {
	var variants []models.Variant
	if len(ids) == 0 {
		return variants, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&variants).Error
	return variants, err
}

func (r *variantRepo) List() ([]models.Variant, error) {
	var variants []models.Variant
	err := r.db.Find(&variants).Error
	return variants, err
}

func (r *variantRepo) Update(v *models.Variant) (*models.Variant, error) {
	err := r.db.Save(v).Error
	return v, err
}

func (r *variantRepo) Patch(id uint, updates map[string]interface{}) (*models.Variant, error) {
	var variant models.Variant
	if err := r.db.First(&variant, id).Error; err != nil {
		return nil, err
	}

	if imgData, ok := updates["images"]; ok {
		r.db.Where("variant_id = ?", id).Delete(&models.ProductImage{})

		if urls, ok := imgData.([]interface{}); ok {
			var newImages []models.ProductImage
			for _, u := range urls {
				if urlStr, ok := u.(string); ok {
					newImages = append(newImages, models.ProductImage{VariantID: id, URL: urlStr})
				}
			}
			if len(newImages) > 0 {
				r.db.Create(&newImages)
			}
		}
		delete(updates, "images")
	}

	if len(updates) > 0 {
		if err := r.db.Model(&variant).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return r.GetByID(id)
}

func (r *variantRepo) Delete(id uint) error {
	return r.db.Delete(&models.Variant{}, id).Error
}

// тут недо жестко попотеть и обновить
func (r *variantRepo) Search(filter models.VariantFilter) ([]models.VariantListItemDTO, error) {
	var results []models.VariantListItemDTO

	query := r.db.Model(&models.Variant{})

	if filter.Name != nil || filter.CategoryID != nil {
		query = query.Joins("JOIN products ON products.id = variants.product_id")
		if filter.Name != nil {
			searchPattern := "%" + strings.ToLower(*filter.Name) + "%"
			query = query.Where("LOWER(products.name) LIKE ? OR LOWER(variants.sku) LIKE ?", searchPattern, searchPattern)
		}
		if filter.CategoryID != nil {
			query = query.Where("products.category_id = ?", *filter.CategoryID)
		}
	}

	if filter.SKU != nil {
		query = query.Where("LOWER(variants.sku) = LOWER(?)", *filter.SKU)
	}

	if filter.StockStatus != "all" {
		if filter.WarehouseID != nil {
			switch filter.StockStatus {
			case "in_stock":
				query = query.Where("EXISTS (SELECT 1 FROM stock_balances sb WHERE sb.item_id = variants.id AND sb.warehouse_id = ? AND sb.quantity > 0)", *filter.WarehouseID)
			case "out_of_stock":
				query = query.Where("NOT EXISTS (SELECT 1 FROM stock_balances sb WHERE sb.item_id = variants.id AND sb.warehouse_id = ? AND sb.quantity > 0)", *filter.WarehouseID)
			}
		} else {
			switch filter.StockStatus {
			case "in_stock":
				query = query.Where("EXISTS (SELECT 1 FROM stock_balances sb WHERE sb.item_id = variants.id AND sb.quantity > 0)")
			case "out_of_stock":
				query = query.Where("NOT EXISTS (SELECT 1 FROM stock_balances sb WHERE sb.item_id = variants.id AND sb.quantity > 0)")
			}
		}
	}

	var variantIDs []uint
	err := query.Order("variants.id asc").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Pluck("variants.id", &variantIDs).Error

	if err != nil {
		return nil, err
	}
	if len(variantIDs) == 0 {
		return []models.VariantListItemDTO{}, nil
	}

	selects := []string{
		"variants.id", "variants.product_id", "products.name as product_name",
		"variants.sku", "variants.characteristics", "variants.unit_id", "units.name as unit_name",
		"products.category_id", "categories.name as category_name",
	}

	enrichQuery := r.db.Model(&models.Variant{}).
		Joins("JOIN products ON products.id = variants.product_id").
		Joins("JOIN units ON units.id = variants.unit_id").
		Joins("JOIN categories ON categories.id = products.category_id").
		Where("variants.id IN ?", variantIDs)

	if filter.WarehouseID != nil {
		selects = append(selects, "COALESCE(sb.quantity, 0) as quantity_on_stock")
		enrichQuery = enrichQuery.Joins(
			"LEFT JOIN stock_balances sb ON sb.item_id = variants.id AND sb.warehouse_id = ?", *filter.WarehouseID,
		)
	} else {
		selects = append(selects, `
			(SELECT COALESCE(SUM(sb.quantity), 0) 
			 FROM stock_balances sb 
			 WHERE sb.item_id = variants.id) as quantity_on_stock
		`)
	}

	enrichQuery = enrichQuery.Select(strings.Join(selects, ", "))

	err = enrichQuery.Order("variants.id asc").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	var images []models.ProductImage
	r.db.Where("variant_id IN ?", variantIDs).Find(&images)

	imgMap := make(map[uint][]models.ProductImageDTO)
	for _, img := range images {
		imgMap[img.VariantID] = append(imgMap[img.VariantID], models.ProductImageDTO{
			ID:  img.ID,
			URL: img.URL,
		})
	}

	for i := range results {
		if imgs, ok := imgMap[results[i].ID]; ok {
			results[i].Images = imgs
		} else {
			results[i].Images = []models.ProductImageDTO{}
		}
	}

	return results, nil
}

func (r *variantRepo) FindByProductID(productID uint) ([]models.Variant, error) {
	var variants []models.Variant
	err := r.db.Where("product_id = ?", productID).Preload("Images").Order("id asc").Find(&variants).Error
	return variants, err
}
