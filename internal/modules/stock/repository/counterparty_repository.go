package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type CounterpartyRepository interface {
	Create(cp *stock.Counterparty) (*stock.Counterparty, error)
	GetByID(id uint) (*stock.Counterparty, error)
	List() ([]stock.Counterparty, error)
	Update(cp *stock.Counterparty) (*stock.Counterparty, error)
	Delete(id uint) error
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
