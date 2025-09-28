package repository

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type VariantRepository interface {
	Create(v *models.Variant) (*models.Variant, error)
	GetByID(id uint) (*models.Variant, error)
	List() ([]models.Variant, error)
	Update(v *models.Variant) (*models.Variant, error)
	Delete(id uint) error
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
