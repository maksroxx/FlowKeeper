package service

import (
	"testing"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/stretchr/testify/assert"
)

type mockCategoryRepo struct {
	repository.CategoryRepository
	db map[uint]*models.Category
}

func (m *mockCategoryRepo) Create(c *models.Category) (*models.Category, error) {
	c.ID = uint(len(m.db) + 1)
	m.db[c.ID] = c
	return c, nil
}

func (m *mockCategoryRepo) List() ([]models.Category, error) {
	var list []models.Category
	for _, c := range m.db {
		list = append(list, *c)
	}
	return list, nil
}

func TestCategoryService_CreateAndList(t *testing.T) {
	repo := &mockCategoryRepo{db: make(map[uint]*models.Category)}
	svc := NewCategoryService(repo)

	t.Run("Создание_и_список", func(t *testing.T) {
		c1, err1 := svc.Create("Электроника")
		c2, err2 := svc.Create("Одежда")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "Электроника", c1.Name)
		assert.Equal(t, "Одежда", c2.Name)

		list, err := svc.List()
		assert.NoError(t, err)
		assert.Len(t, list, 2)
	})
}
