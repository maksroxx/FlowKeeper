package service

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type PriceTypeService interface {
	Create(name string) (*models.PriceType, error)
	GetByID(id uint) (*models.PriceType, error)
	List() ([]models.PriceType, error)
	Update(pt *models.PriceType) (*models.PriceType, error)
	Delete(id uint) error
}

type priceTypeService struct {
	repo repository.PriceTypeRepository
}

func NewPriceTypeService(r repository.PriceTypeRepository) PriceTypeService {
	return &priceTypeService{repo: r}
}

func (s *priceTypeService) Create(name string) (*models.PriceType, error) {
	return s.repo.Create(&models.PriceType{Name: name})
}
func (s *priceTypeService) GetByID(id uint) (*models.PriceType, error) {
	return s.repo.GetByID(id)
}
func (s *priceTypeService) List() ([]models.PriceType, error) {
	return s.repo.List()
}
func (s *priceTypeService) Update(pt *models.PriceType) (*models.PriceType, error) {
	return s.repo.Update(pt)
}
func (s *priceTypeService) Delete(id uint) error {
	return s.repo.Delete(id)
}
