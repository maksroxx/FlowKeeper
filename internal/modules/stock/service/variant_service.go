package service

import (
	"encoding/json"
	"errors"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type VariantService interface {
	Create(v *models.Variant, images []string) (*models.Variant, error)
	GetByID(id uint) (*models.Variant, error)
	GetByIDAsDTO(id uint) (*models.VariantDTO, error)
	List() ([]models.Variant, error)
	Update(id uint, updates map[string]interface{}) (*models.Variant, error)
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

func (s *variantService) Create(v *models.Variant, images []string) (*models.Variant, error) {
	for _, url := range images {
		v.Images = append(v.Images, models.ProductImage{URL: url})
	}
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

// func (s *variantService) Update(id uint, updateData *models.Variant) (*models.Variant, error) {
// 	variantToUpdate, err := s.repo.GetByID(id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if variantToUpdate == nil {
// 		return nil, errors.New("variant not found to update")
// 	}

// 	variantToUpdate.SKU = updateData.SKU
// 	variantToUpdate.UnitID = updateData.UnitID
// 	variantToUpdate.Characteristics = updateData.Characteristics
// 	variantToUpdate.ProductID = updateData.ProductID

// 	return s.repo.Update(variantToUpdate)
// }

func (s *variantService) Update(id uint, updates map[string]interface{}) (*models.Variant, error) {
	delete(updates, "product_id")
	if chars, ok := updates["characteristics"]; ok {
		jsonBytes, err := json.Marshal(chars)
		if err != nil {
			return nil, errors.New("invalid format for characteristics")
		}
		updates["characteristics"] = string(jsonBytes)
	}

	return s.repo.Patch(id, updates)
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

	for _, img := range variant.Images {
		dto.Images = append(dto.Images, models.ProductImageDTO{ID: img.ID, URL: img.URL})
	}

	if product, err := s.productRepo.GetByID(variant.ProductID); err == nil && product != nil {
		dto.ProductName = product.Name
	}
	if unit, err := s.unitRepo.GetByID(variant.UnitID); err == nil && unit != nil {
		dto.UnitName = unit.Name
	}

	return dto, nil
}
