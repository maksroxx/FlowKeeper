package repository

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type PriceTypeRepository interface {
	Create(pt *models.PriceType) (*models.PriceType, error)
	GetByID(id uint) (*models.PriceType, error)
	List() ([]models.PriceType, error)
	Update(pt *models.PriceType) (*models.PriceType, error)
	Delete(id uint) error
}

type priceTypeRepo struct{ db *gorm.DB }

func NewPriceTypeRepository(db *gorm.DB) PriceTypeRepository {
	return &priceTypeRepo{db: db}
}

func (r *priceTypeRepo) Create(pt *models.PriceType) (*models.PriceType, error) {
	err := r.db.Create(pt).Error
	return pt, err
}

func (r *priceTypeRepo) GetByID(id uint) (*models.PriceType, error) {
	var pt models.PriceType
	if err := r.db.First(&pt, id).Error; err != nil {
		return nil, err
	}
	return &pt, nil
}

func (r *priceTypeRepo) List() ([]models.PriceType, error) {
	var pts []models.PriceType
	err := r.db.Find(&pts).Error
	return pts, err
}

func (r *priceTypeRepo) Update(pt *models.PriceType) (*models.PriceType, error) {
	err := r.db.Save(pt).Error
	return pt, err
}

func (r *priceTypeRepo) Delete(id uint) error {
	return r.db.Delete(&models.PriceType{}, id).Error
}
