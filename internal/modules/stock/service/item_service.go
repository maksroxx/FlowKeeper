package service

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type VariantService interface {
	Create(v *models.Variant) (*models.Variant, error)
	GetByID(id uint) (*models.Variant, error)
	GetByIDAsDTO(id uint) (*models.VariantDTO, error)
	List() ([]models.Variant, error)
	Update(v *models.Variant) (*models.Variant, error)
	Delete(id uint) error
	Search(filter models.VariantFilter) ([]models.VariantListItemDTO, error)
}

type variantService struct {
	repo        repository.VariantRepository
	productRepo repository.ProductRepository
	unitRepo    repository.UnitRepository
}

func NewVariantService(
	repo repository.VariantRepository,
	productRepo repository.ProductRepository,
	unitRepo repository.UnitRepository,
) VariantService {
	return &variantService{repo: repo, productRepo: productRepo, unitRepo: unitRepo}
}

func (s *variantService) Create(v *models.Variant) (*models.Variant, error) {
	return s.repo.Create(v)
}

func (s *variantService) GetByID(id uint) (*models.Variant, error) {
	return s.repo.GetByID(id)
}

func (s *variantService) GetByIDAsDTO(id uint) (*models.VariantDTO, error) {
	variant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, nil
	}
	return s.buildDTO(variant)
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

func (s *variantService) Search(filter models.VariantFilter) ([]models.VariantListItemDTO, error) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.StockStatus == "" {
		filter.StockStatus = "all"
	}

	return s.repo.Search(filter)
}

func (s *variantService) buildDTO(variant *models.Variant) (*models.VariantDTO, error) {
	dto := &models.VariantDTO{
		ID:              variant.ID,
		ProductID:       variant.ProductID,
		SKU:             variant.SKU,
		Characteristics: variant.Characteristics,
		UnitID:          variant.UnitID,
	}

	if product, err := s.productRepo.GetByID(variant.ProductID); err == nil && product != nil {
		dto.ProductName = product.Name
	}
	if unit, err := s.unitRepo.GetByID(variant.UnitID); err == nil && unit != nil {
		dto.UnitName = unit.Name
	}

	return dto, nil
}
