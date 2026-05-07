package service

import (
	"testing"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/stretchr/testify/assert"
)

type mockProductRepo struct {
	repository.ProductRepository
}

func (m *mockProductRepo) Create(p *models.Product) (*models.Product, error) {
	p.ID = 100
	return p, nil
}

type mockVariantRepo struct {
	repository.VariantRepository
	createdVariant *models.Variant
}

func (m *mockVariantRepo) Create(v *models.Variant) (*models.Variant, error) {
	v.ID = 500
	m.createdVariant = v
	return v, nil
}

func TestProductService_Create(t *testing.T) {
	pRepo := &mockProductRepo{}
	vRepo := &mockVariantRepo{}
	svc := NewProductService(pRepo, vRepo)

	t.Run("Товар_с_вариантом", func(t *testing.T) {
		prod := &models.Product{
			Name:       "iPhone 15",
			CategoryID: 1,
		}
		chars := map[string]string{"Цвет": "Черный", "Память": "256GB"}
		images := []string{"http://img1.jpg", "http://img2.jpg"}

		createdProd, err := svc.Create(prod, "IPH-15-BLK", 2, chars, images)

		assert.NoError(t, err)
		assert.Equal(t, uint(100), createdProd.ID)

		assert.NotNil(t, vRepo.createdVariant)
		assert.Equal(t, uint(100), vRepo.createdVariant.ProductID)
		assert.Equal(t, "IPH-15-BLK", vRepo.createdVariant.SKU)
		assert.Equal(t, uint(2), vRepo.createdVariant.UnitID)
		assert.Equal(t, "Черный", vRepo.createdVariant.Characteristics["Цвет"])

		assert.Len(t, vRepo.createdVariant.Images, 2)
		assert.Equal(t, "http://img1.jpg", vRepo.createdVariant.Images[0].URL)
	})
}
