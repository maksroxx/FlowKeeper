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

func (r *productRepo) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}
