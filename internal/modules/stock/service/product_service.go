package service

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type ProductService interface {
	Create(p *models.Product) (*models.Product, error)
	GetByID(id uint) (*models.Product, error)
	List() ([]models.Product, error)
	Update(p *models.Product) (*models.Product, error)
	Delete(id uint) error
}

type productService struct{ repo repository.ProductRepository }

func NewProductService(r repository.ProductRepository) ProductService {
	return &productService{repo: r}
}

func (s *productService) Create(p *models.Product) (*models.Product, error) {
	return s.repo.Create(p)
}

func (s *productService) GetByID(id uint) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *productService) List() ([]models.Product, error) {
	return s.repo.List()
}

func (s *productService) Update(p *models.Product) (*models.Product, error) {
	return s.repo.Update(p)
}

func (s *productService) Delete(id uint) error {
	return s.repo.Delete(id)
}
