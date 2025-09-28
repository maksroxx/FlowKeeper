package service

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type VariantService interface {
	Create(v *models.Variant) (*models.Variant, error)
	GetByID(id uint) (*models.Variant, error)
	List() ([]models.Variant, error)
	Update(v *models.Variant) (*models.Variant, error)
	Delete(id uint) error
}

type variantService struct{ repo repository.VariantRepository }

func NewVariantService(r repository.VariantRepository) VariantService {
	return &variantService{repo: r}
}

func (s *variantService) Create(v *models.Variant) (*models.Variant, error) {
	return s.repo.Create(v)
}

func (s *variantService) GetByID(id uint) (*models.Variant, error) {
	return s.repo.GetByID(id)
}

func (s *variantService) List() ([]models.Variant, error) {
	return s.repo.List()
}

func (s *variantService) Update(v *models.Variant) (*models.Variant, error) {
	return s.repo.Update(v)
}

func (s *variantService) Delete(id uint) error {
	return s.repo.Delete(id)
}
