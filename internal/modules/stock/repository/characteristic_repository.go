package repository

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type CharacteristicRepository interface {
	CreateType(ct *models.CharacteristicType) (*models.CharacteristicType, error)
	GetTypeByID(id uint) (*models.CharacteristicType, error)
	ListTypes() ([]models.CharacteristicType, error)
	UpdateType(ct *models.CharacteristicType) (*models.CharacteristicType, error)
	DeleteType(id uint) error

	CreateValue(cv *models.CharacteristicValue) (*models.CharacteristicValue, error)
	GetValueByID(id uint) (*models.CharacteristicValue, error)
	ListValues() ([]models.CharacteristicValue, error)
	UpdateValue(cv *models.CharacteristicValue) (*models.CharacteristicValue, error)
	DeleteValue(id uint) error
}

type characteristicRepo struct{ db *gorm.DB }

func NewCharacteristicRepository(db *gorm.DB) CharacteristicRepository {
	return &characteristicRepo{db: db}
}

func (r *characteristicRepo) CreateType(ct *models.CharacteristicType) (*models.CharacteristicType, error) {
	err := r.db.Create(ct).Error
	return ct, err
}

func (r *characteristicRepo) GetTypeByID(id uint) (*models.CharacteristicType, error) {
	var ct models.CharacteristicType
	if err := r.db.First(&ct, id).Error; err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *characteristicRepo) ListTypes() ([]models.CharacteristicType, error) {
	var cts []models.CharacteristicType
	err := r.db.Find(&cts).Error
	return cts, err
}

func (r *characteristicRepo) UpdateType(ct *models.CharacteristicType) (*models.CharacteristicType, error) {
	err := r.db.Save(ct).Error
	return ct, err
}

func (r *characteristicRepo) DeleteType(id uint) error {
	return r.db.Delete(&models.CharacteristicType{}, id).Error
}

func (r *characteristicRepo) CreateValue(cv *models.CharacteristicValue) (*models.CharacteristicValue, error) {
	err := r.db.Create(cv).Error
	return cv, err
}

func (r *characteristicRepo) GetValueByID(id uint) (*models.CharacteristicValue, error) {
	var cv models.CharacteristicValue
	if err := r.db.First(&cv, id).Error; err != nil {
		return nil, err
	}
	return &cv, nil
}

func (r *characteristicRepo) ListValues() ([]models.CharacteristicValue, error) {
	var cvs []models.CharacteristicValue
	err := r.db.Find(&cvs).Error
	return cvs, err
}

func (r *characteristicRepo) UpdateValue(cv *models.CharacteristicValue) (*models.CharacteristicValue, error) {
	err := r.db.Save(cv).Error
	return cv, err
}

func (r *characteristicRepo) DeleteValue(id uint) error {
	return r.db.Delete(&models.CharacteristicValue{}, id).Error
}
