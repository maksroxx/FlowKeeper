package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWarehouse_CRUD_Integration(t *testing.T) {
	router, db := setupTestRouter("warehouse_crud_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	createdWH := h.CreateWarehouse("Центральный склад")
	h.Assert.Equal("Центральный склад", createdWH.Name)
	h.Assert.NotZero(createdWH.ID)

	fetchedWH := h.GetWarehouse(createdWH.ID)
	h.Assert.Equal(createdWH.Name, fetchedWH.Name)

	updatedWH := h.UpdateWarehouse(createdWH.ID, "Новое название склада")
	h.Assert.Equal("Новое название склада", updatedWH.Name)

	allWHs := h.ListWarehouses()
	h.Assert.Len(allWHs, 1)
	h.Assert.Equal("Новое название склада", allWHs[0].Name)

	h.DeleteWarehouse(createdWH.ID)

	finalWHs := h.ListWarehouses()
	h.Assert.Len(finalWHs, 0)
}

func TestProductAndVariant_Creation_Integration(t *testing.T) {
	router, db := setupTestRouter("product_variant_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	category := h.CreateCategory("Футболки")
	unit := h.CreateUnit("шт")

	product := h.CreateProduct(gin.H{
		"name":        "Футболка поло",
		"category_id": category.ID,
	})

	variant := h.CreateVariant(gin.H{
		"product_id": product.ID,
		"sku":        "POLO-BLUE-L",
		"unit_id":    unit.ID,
		"characteristics": gin.H{
			"Цвет":   "Синий",
			"Размер": "L",
		},
	})

	h.Assert.NotZero(product.ID, "Product ID should not be zero after creation")
	h.Assert.Equal(product.ID, variant.ProductID)
	h.Assert.Equal("POLO-BLUE-L", variant.SKU)
	h.Assert.Equal("Синий", variant.Characteristics["Цвет"])
}
