package repository

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(p *models.Product) (*models.Product, error)
	GetByID(id uint) (*models.Product, error)
	GetByIDs(ids []uint) ([]models.Product, error)
	List() ([]models.Product, error)
	Update(p *models.Product) (*models.Product, error)
	Patch(id uint, updates map[string]interface{}) (*models.Product, error)
	Delete(id uint) error

	ImportBatch(items []models.ImportItemDTO) error
}

type productRepo struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(p *models.Product) (*models.Product, error) {
	err := r.db.Create(p).Error
	return p, err
}

func (r *productRepo) GetByID(id uint) (*models.Product, error) {
	var p models.Product
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepo) GetByIDs(ids []uint) ([]models.Product, error) {
	var products []models.Product
	if len(ids) == 0 {
		return products, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&products).Error
	return products, err
}

func (r *productRepo) List() ([]models.Product, error) {
	var ps []models.Product
	err := r.db.Find(&ps).Error
	return ps, err
}

func (r *productRepo) Update(p *models.Product) (*models.Product, error) {
	err := r.db.Save(p).Error
	return p, err
}

func (r *productRepo) Patch(id uint, updates map[string]interface{}) (*models.Product, error) {
	var product models.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&product).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *productRepo) ImportBatch(items []models.ImportItemDTO) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var cat models.Category
			if err := tx.Where(models.Category{Name: item.CategoryName}).
				FirstOrCreate(&cat).Error; err != nil {
				return err
			}

			var unit models.Unit
			if err := tx.Where(models.Unit{Name: item.UnitName}).
				FirstOrCreate(&unit).Error; err != nil {
				return err
			}

			var product models.Product
			if err := tx.Where("name = ?", item.ProductName).First(&product).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					product = models.Product{
						Name:        item.ProductName,
						Description: item.Description,
						CategoryID:  cat.ID,
					}
					if err := tx.Create(&product).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				product.CategoryID = cat.ID
				tx.Save(&product)
			}

			var variant models.Variant
			if err := tx.Where("sku = ?", item.SKU).First(&variant).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					variant = models.Variant{
						ProductID:       product.ID,
						SKU:             item.SKU,
						UnitID:          unit.ID,
						Characteristics: item.Characteristics,
					}
					if err := tx.Create(&variant).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				variant.Characteristics = item.Characteristics
				tx.Save(&variant)
			}
		}
		return nil
	})
}
