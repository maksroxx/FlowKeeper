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
	if err := r.db.Preload("Images").First(&p, id).Error; err != nil {
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
	err := r.db.Preload("Images").Find(&ps).Error
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

	if imgData, ok := updates["images"]; ok {
		r.db.Where("product_id = ?", id).Delete(&models.ProductImage{})
		if urls, ok := imgData.([]interface{}); ok {
			var newImages []models.ProductImage
			for _, u := range urls {
				if urlStr, ok := u.(string); ok {
					newImages = append(newImages, models.ProductImage{ProductID: id, URL: urlStr})
				}
			}
			r.db.Create(&newImages)
		}
		delete(updates, "images")
	}

	if len(updates) > 0 {
		if err := r.db.Model(&product).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return r.GetByID(id)
}

func (r *productRepo) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}
