package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestDirectoryManagement_FullCycle(t *testing.T) {
	// 1. ARRANGE
	router, db := setupTestRouter("dir_mgmt_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	// Создаем базовые справочники
	category := h.CreateCategory("Электроника")
	unit := h.CreateUnit("шт")

	// 2. ACT & ASSERT: Работа с Номенклатурой (Product)
	t.Log("--- Тестирование CRUD для Product ---")

	// Create
	product := h.CreateProduct(gin.H{"name": "Смартфон Model Z", "category_id": category.ID})
	h.Assert.Equal("Смартфон Model Z", product.Name)
	h.Assert.NotZero(product.ID)

	// Get by ID
	fetchedProduct := h.GetProduct(product.ID)
	h.Assert.Equal(product.Name, fetchedProduct.Name)

	// List
	productList := h.ListProducts()
	h.Assert.Len(productList, 1)

	// Update
	updatedProduct := h.UpdateProduct(product.ID, gin.H{"name": "Смартфон Model Z Pro", "description": "Флагманская модель"})
	h.Assert.Equal("Смартфон Model Z Pro", updatedProduct.Name)
	h.Assert.Equal("Флагманская модель", updatedProduct.Description)

	// 3. ACT & ASSERT: Работа с Вариантами (Variant)
	t.Log("--- Тестирование CRUD для Variant ---")

	// Create Variant 1
	variant1 := h.CreateVariant(gin.H{
		"product_id":      updatedProduct.ID,
		"sku":             "Z-PRO-BLK-128",
		"unit_id":         unit.ID,
		"characteristics": gin.H{"Цвет": "Черный", "Память": "128Гб"},
	})
	h.Assert.Equal("Z-PRO-BLK-128", variant1.SKU)

	// Create Variant 2
	variant2 := h.CreateVariant(gin.H{
		"product_id":      updatedProduct.ID,
		"sku":             "Z-PRO-WHT-256",
		"unit_id":         unit.ID,
		"characteristics": gin.H{"Цвет": "Белый", "Память": "256Гб"},
	})

	// Get by ID (для Variant)
	fetchedVariant := h.GetVariant(variant2.ID)
	h.Assert.Equal("Z-PRO-WHT-256", fetchedVariant.SKU)

	// Search (List all variants)
	variantList := h.SearchVariants("")
	h.Assert.Len(variantList, 2, "Должно быть 2 варианта у продукта")

	// Update Variant
	updatedVariant := h.UpdateVariant(variant1.ID, gin.H{
		"product_id":      variant1.ProductID,
		"unit_id":         variant1.UnitID,
		"sku":             "Z-PRO-BLACK-128-NEW",
		"characteristics": variant1.Characteristics,
	})
	h.Assert.Equal("Z-PRO-BLACK-128-NEW", updatedVariant.SKU)

	// Delete Variant
	h.DeleteVariant(variant2.ID)
	variantListAfterDelete := h.SearchVariants("")
	h.Assert.Len(variantListAfterDelete, 1, "Должен остаться 1 вариант после удаления")

	// 4. ACT & ASSERT: Работа с Документами (проверка CRUD)
	t.Log("--- Тестирование CRUD для Document ---")

	// Create Document
	doc := h.CreateDocument(models.Document{
		Type:  "DRAFT_TEST",
		Items: []models.DocumentItem{{VariantID: variant1.ID}},
	})
	h.Assert.NotZero(doc.ID)
	h.Assert.Equal("draft", doc.Status) // Проверяем статус по умолчанию

	// Get by ID Document
	fetchedDoc := h.GetDocument(doc.ID)
	h.Assert.Equal("DRAFT_TEST", fetchedDoc.Type)

	// Update Document
	updatedDoc := h.UpdateDocument(doc.ID, gin.H{"comment": "Тестовый комментарий", "type": "DRAFT_TEST"})
	h.Assert.Equal("Тестовый комментарий", updatedDoc.Comment)

	// ИСПРАВЛЕНО: Ищем документы по их реальному статусу - "draft"
	docList := h.ListDocuments("draft")

	// Убеждаемся, что наш документ есть в списке
	h.Assert.GreaterOrEqual(len(docList), 1, "Список документов со статусом 'draft' не должен быть пустым")
	found := false
	for _, d := range docList {
		if d.ID == doc.ID {
			found = true
			break
		}
	}
	h.Assert.True(found, "Наш тестовый документ должен был быть найден в списке черновиков")

	// Delete Document
	h.DeleteDocument(doc.ID)

	// Проверяем, что после удаления его больше нет в списке
	docListAfterDelete := h.ListDocuments("draft")
	foundAfterDelete := false
	for _, d := range docListAfterDelete {
		if d.ID == doc.ID {
			foundAfterDelete = true
			break
		}
	}
	h.Assert.False(foundAfterDelete, "Наш тестовый документ не должен был быть найден после удаления")

	// 5. ACT & ASSERT: Финальная очистка
	t.Log("--- Тестирование финальной очистки ---")

	// Сначала удаляем оставшийся вариант
	h.DeleteVariant(variant1.ID)
	// Затем удаляем родительский продукт
	h.DeleteProduct(product.ID)

	finalProductList := h.ListProducts()
	h.Assert.Len(finalProductList, 0, "Все продукты должны быть удалены")
}
