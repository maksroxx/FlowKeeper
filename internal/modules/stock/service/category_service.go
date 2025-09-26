package service

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type CategoryService interface {
	Create(name string) (*stock.Category, error)
	GetByID(id uint) (*stock.Category, error)
	List() ([]stock.Category, error)
	Update(c *stock.Category) (*stock.Category, error)
	Delete(id uint) error
}

type categoryService struct{ repo repository.CategoryRepository }

func NewCategoryService(r repository.CategoryRepository) CategoryService {
	return &categoryService{repo: r}
}

func (s *categoryService) Create(name string) (*stock.Category, error) {
	return s.repo.Create(&stock.Category{Name: name})
}
func (s *categoryService) GetByID(id uint) (*stock.Category, error) { return s.repo.GetByID(id) }
func (s *categoryService) List() ([]stock.Category, error)          { return s.repo.List() }
func (s *categoryService) Update(c *stock.Category) (*stock.Category, error) {
	return s.repo.Update(c)
}
func (s *categoryService) Delete(id uint) error { return s.repo.Delete(id) }
