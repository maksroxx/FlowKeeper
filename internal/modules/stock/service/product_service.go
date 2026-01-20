package service

import (
	"errors"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type ProductService interface {
	Create(p *models.Product, sku string, unitID uint, characteristics map[string]string, images []string) (*models.Product, error)
	GetByID(id uint) (*models.Product, error)
	List() ([]models.Product, error)
	Update(id uint, updates map[string]interface{}) (*models.Product, error)
	Delete(id uint) error

	GetProductDetails(productID uint) (*models.ProductDetailDTO, error)
}

type productService struct {
	repo        repository.ProductRepository
	variantRepo repository.VariantRepository
}

func NewProductService(repo repository.ProductRepository, variantRepo repository.VariantRepository) ProductService {
	return &productService{repo: repo, variantRepo: variantRepo}
}

func (s *productService) Create(p *models.Product, sku string, unitID uint, char map[string]string, images []string) (*models.Product, error) {
	createdProduct, err := s.repo.Create(p)
	if err != nil {
		return nil, err
	}

	v := &models.Variant{
		ProductID:       createdProduct.ID,
		SKU:             sku,
		UnitID:          unitID,
		Characteristics: char,
	}

	for _, url := range images {
		v.Images = append(v.Images, models.ProductImage{URL: url})
	}

	if _, err := s.variantRepo.Create(v); err != nil {
		return nil, err
	}

	return createdProduct, nil
}

func (s *productService) GetByID(id uint) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *productService) List() ([]models.Product, error) {
	return s.repo.List()
}

// func (s *productService) Update(id uint, updateData *models.Product) (*models.Product, error) {
// 	productToUpdate, err := s.repo.GetByID(id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if productToUpdate == nil {
// 		return nil, errors.New("product not found to update")
// 	}

// 	productToUpdate.Name = updateData.Name
// 	productToUpdate.Description = updateData.Description
// 	productToUpdate.CategoryID = updateData.CategoryID

// 	return s.repo.Update(productToUpdate)
// }

func (s *productService) Update(id uint, updates map[string]interface{}) (*models.Product, error) {
	return s.repo.Patch(id, updates)
}

func (s *productService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *productService) GetProductDetails(productID uint) (*models.ProductDetailDTO, error) {
	product, err := s.repo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	variants, err := s.variantRepo.FindByProductID(productID)
	if err != nil {
		return nil, err
	}

	optionsMap := make(map[string]map[string]bool)
	for _, v := range variants {
		for charType, charValue := range v.Characteristics {
			if _, ok := optionsMap[charType]; !ok {
				optionsMap[charType] = make(map[string]bool)
			}
			optionsMap[charType][charValue] = true
		}
	}

	options := make([]models.ProductOptionDTO, 0, len(optionsMap))
	for charType, valuesMap := range optionsMap {
		values := make([]string, 0, len(valuesMap))
		for value := range valuesMap {
			values = append(values, value)
		}
		options = append(options, models.ProductOptionDTO{Type: charType, Values: values})
	}

	details := &models.ProductDetailDTO{
		Product:  *product,
		Variants: variants,
		Options:  options,
	}

	return details, nil
}
