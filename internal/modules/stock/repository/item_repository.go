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
	Delete(id uint) error
	Search(filter models.VariantFilter) ([]models.VariantListItemDTO, error)
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
	if err := r.db.First(&variant, id).Error; err != nil {
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

func (r *variantRepo) Delete(id uint) error {
	return r.db.Delete(&models.Variant{}, id).Error
}

func (r *variantRepo) Search(filter models.VariantFilter) ([]models.VariantListItemDTO, error) {
	var results []models.VariantListItemDTO

	query := r.db.Model(&models.Variant{})

	selects := []string{
		"variants.id",
		"variants.product_id",
		"products.name as product_name",
		"variants.sku",
		"variants.characteristics",
		"variants.unit_id",
		"units.name as unit_name",
	}

	query = query.Joins("JOIN products ON products.id = variants.product_id")
	query = query.Joins("JOIN units ON units.id = variants.unit_id")

	if filter.WarehouseID != nil {
		selects = append(selects, "COALESCE(sb.quantity, 0) as quantity_on_stock")
		query = query.Joins("LEFT JOIN stock_balances sb ON sb.item_id = variants.id AND sb.warehouse_id = ?", *filter.WarehouseID)
		switch filter.StockStatus {
		case "in_stock":
			query = query.Where("sb.quantity > 0")
		case "out_of_stock":
			query = query.Where("sb.quantity IS NULL OR sb.quantity <= 0")
		}
	} else {
		selects = append(selects, "0 as quantity_on_stock")
	}

	query = query.Select(strings.Join(selects, ", "))

	if filter.SKU != nil {
		query = query.Where("LOWER(variants.sku) = LOWER(?)", *filter.SKU)
	}
	if filter.Name != nil {
		searchPattern := "%" + strings.ToLower(*filter.Name) + "%"
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(variants.sku) LIKE ?", searchPattern, searchPattern)
	}
	if filter.CategoryID != nil {
		query = query.Where("products.category_id = ?", *filter.CategoryID)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("variants.id asc").Scan(&results).Error
	return results, err
}
