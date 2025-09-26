package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type UnitRepository interface {
	Create(u *stock.Unit) (*stock.Unit, error)
	GetByID(id uint) (*stock.Unit, error)
	List() ([]stock.Unit, error)
	Update(u *stock.Unit) (*stock.Unit, error)
	Delete(id uint) error
}

type unitRepo struct{ db *gorm.DB }

func NewUnitRepository(db *gorm.DB) UnitRepository { return &unitRepo{db: db} }

func (r *unitRepo) Create(u *stock.Unit) (*stock.Unit, error) {
	err := r.db.Create(u).Error
	return u, err
}

func (r *unitRepo) GetByID(id uint) (*stock.Unit, error) {
	var unit stock.Unit
	if err := r.db.First(&unit, id).Error; err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *unitRepo) List() ([]stock.Unit, error) {
	var units []stock.Unit
	err := r.db.Find(&units).Error
	return units, err
}

func (r *unitRepo) Update(u *stock.Unit) (*stock.Unit, error) {
	err := r.db.Save(u).Error
	return u, err
}

func (r *unitRepo) Delete(id uint) error {
	return r.db.Delete(&stock.Unit{}, id).Error
}
