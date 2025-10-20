package service

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type CharacteristicService interface {
	GetTree() ([]models.CharacteristicTreeDTO, error)

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

type characteristicService struct {
	repo repository.CharacteristicRepository
}

func NewCharacteristicService(r repository.CharacteristicRepository) CharacteristicService {
	return &characteristicService{repo: r}
}

func (s *characteristicService) CreateType(ct *models.CharacteristicType) (*models.CharacteristicType, error) {
	return s.repo.CreateType(ct)
}

func (s *characteristicService) GetTypeByID(id uint) (*models.CharacteristicType, error) {
	return s.repo.GetTypeByID(id)
}

func (s *characteristicService) ListTypes() ([]models.CharacteristicType, error) {
	return s.repo.ListTypes()
}

func (s *characteristicService) UpdateType(ct *models.CharacteristicType) (*models.CharacteristicType, error) {
	return s.repo.UpdateType(ct)
}

func (s *characteristicService) DeleteType(id uint) error {
	return s.repo.DeleteType(id)
}

func (s *characteristicService) CreateValue(cv *models.CharacteristicValue) (*models.CharacteristicValue, error) {
	return s.repo.CreateValue(cv)
}

func (s *characteristicService) GetValueByID(id uint) (*models.CharacteristicValue, error) {
	return s.repo.GetValueByID(id)
}

func (s *characteristicService) ListValues() ([]models.CharacteristicValue, error) {
	return s.repo.ListValues(nil)
}

func (s *characteristicService) UpdateValue(cv *models.CharacteristicValue) (*models.CharacteristicValue, error) {
	return s.repo.UpdateValue(cv)
}

func (s *characteristicService) DeleteValue(id uint) error {
	return s.repo.DeleteValue(id)
}

func (s *characteristicService) GetTree() ([]models.CharacteristicTreeDTO, error) {
	types, err := s.repo.ListTypes()
	if err != nil {
		return nil, err
	}
	if len(types) == 0 {
		return []models.CharacteristicTreeDTO{}, nil
	}

	values, err := s.repo.ListValues(nil)
	if err != nil {
		return nil, err
	}

	valuesMap := make(map[uint][]models.CharacteristicValueDTO)
	for _, v := range values {
		dto := models.CharacteristicValueDTO{ID: v.ID, Value: v.Value}
		valuesMap[v.CharacteristicTypeID] = append(valuesMap[v.CharacteristicTypeID], dto)
	}

	tree := make([]models.CharacteristicTreeDTO, len(types))
	for i, t := range types {
		vals := valuesMap[t.ID]
		if vals == nil {
			vals = []models.CharacteristicValueDTO{}
		}
		tree[i] = models.CharacteristicTreeDTO{
			ID:     t.ID,
			Name:   t.Name,
			Values: vals,
		}
	}
	return tree, nil
}
