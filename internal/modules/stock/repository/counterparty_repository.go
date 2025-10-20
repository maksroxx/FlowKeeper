package repository

import (
	"strings"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type CounterpartyRepository interface {
	Create(cp *stock.Counterparty) (*stock.Counterparty, error)
	GetByID(id uint) (*stock.Counterparty, error)
	GetByIDs(ids []uint) ([]stock.Counterparty, error)
	List() ([]stock.Counterparty, error)
	Update(cp *stock.Counterparty) (*stock.Counterparty, error)
	Delete(id uint) error
	Patch(id uint, updates map[string]interface{}) (*stock.Counterparty, error)
	Search(filter stock.CounterpartyFilter) ([]stock.Counterparty, error)
}

type counterpartyRepository struct{ db *gorm.DB }

func NewCounterpartyRepository(db *gorm.DB) CounterpartyRepository {
	return &counterpartyRepository{db: db}
}

func (r *counterpartyRepository) Create(cp *stock.Counterparty) (*stock.Counterparty, error) {
	err := r.db.Create(cp).Error
	return cp, err
}

func (r *counterpartyRepository) GetByID(id uint) (*stock.Counterparty, error) {
	var cp stock.Counterparty
	if err := r.db.First(&cp, id).Error; err != nil {
		return nil, err
	}
	return &cp, nil
}

func (r *counterpartyRepository) GetByIDs(ids []uint) ([]stock.Counterparty, error) {
	var counterparties []stock.Counterparty
	if len(ids) == 0 {
		return counterparties, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&counterparties).Error
	return counterparties, err
}

func (r *counterpartyRepository) List() ([]stock.Counterparty, error) {
	var cps []stock.Counterparty
	err := r.db.Find(&cps).Error
	return cps, err
}

func (r *counterpartyRepository) Update(cp *stock.Counterparty) (*stock.Counterparty, error) {
	err := r.db.Save(cp).Error
	return cp, err
}

func (r *counterpartyRepository) Delete(id uint) error {
	return r.db.Delete(&stock.Counterparty{}, id).Error
}

func (r *counterpartyRepository) Search(filter stock.CounterpartyFilter) ([]stock.Counterparty, error) {
	var counterparties []stock.Counterparty
	query := r.db.Model(&stock.Counterparty{})

	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + strings.ToLower(*filter.Search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("name asc").Find(&counterparties).Error
	return counterparties, err
}

func (r *counterpartyRepository) Patch(id uint, updates map[string]interface{}) (*stock.Counterparty, error) {
	if err := r.db.Model(&stock.Counterparty{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetByID(id)
}
