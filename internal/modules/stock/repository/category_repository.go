package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(c *stock.Category) (*stock.Category, error)
	GetByID(id uint) (*stock.Category, error)
	GetByIDs(ids []uint) ([]stock.Category, error)
	List() ([]stock.Category, error)
	Update(c *stock.Category) (*stock.Category, error)
	Delete(id uint) error
}

type categoryRepo struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) CategoryRepository { return &categoryRepo{db: db} }

func (r *categoryRepo) Create(c *stock.Category) (*stock.Category, error) {
	err := r.db.Create(c).Error
	return c, err
}

func (r *categoryRepo) GetByID(id uint) (*stock.Category, error) {
	var cat stock.Category
	if err := r.db.First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepo) GetByIDs(ids []uint) ([]stock.Category, error) {
	var categories []stock.Category
	if len(ids) == 0 {
		return categories, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) List() ([]stock.Category, error) {
	var cats []stock.Category
	err := r.db.Find(&cats).Error
	return cats, err
}

func (r *categoryRepo) Update(c *stock.Category) (*stock.Category, error) {
	err := r.db.Save(c).Error
	return c, err
}

func (r *categoryRepo) Delete(id uint) error {
	return r.db.Delete(&stock.Category{}, id).Error
}
