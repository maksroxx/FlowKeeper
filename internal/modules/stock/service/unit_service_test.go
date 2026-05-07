package service

import (
	"errors"
	"testing"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/stretchr/testify/assert"
)

type mockUnitRepo struct {
	repository.UnitRepository
	db map[uint]*models.Unit
}

func (m *mockUnitRepo) Create(u *models.Unit) (*models.Unit, error) {
	if u.Name == "ErrorUnit" {
		return nil, errors.New("db error")
	}
	u.ID = uint(len(m.db) + 1)
	m.db[u.ID] = u
	return u, nil
}

func (m *mockUnitRepo) GetByID(id uint) (*models.Unit, error) {
	if u, ok := m.db[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}

func TestUnitService_Create(t *testing.T) {
	repo := &mockUnitRepo{db: make(map[uint]*models.Unit)}
	svc := NewUnitService(repo)

	t.Run("Успешное_создание", func(t *testing.T) {
		unit, err := svc.Create("шт")
		assert.NoError(t, err)
		assert.Equal(t, "шт", unit.Name)
		assert.NotZero(t, unit.ID)
	})

	t.Run("Ошибка_БД", func(t *testing.T) {
		unit, err := svc.Create("ErrorUnit")
		assert.Error(t, err)
		assert.Nil(t, unit)
	})
}

func TestUnitService_GetByID(t *testing.T) {
	repo := &mockUnitRepo{db: make(map[uint]*models.Unit)}
	svc := NewUnitService(repo)

	created, _ := svc.Create("кг")

	t.Run("Поиск_по_ID_Успех", func(t *testing.T) {
		unit, err := svc.GetByID(created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "кг", unit.Name)
	})

	t.Run("Поиск_по_ID_Ошибка", func(t *testing.T) {
		unit, err := svc.GetByID(999)
		assert.Error(t, err)
		assert.Nil(t, unit)
	})
}
