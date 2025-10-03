package stocktest

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestVariantSearch_Integration(t *testing.T) {
	// 1. ARRANGE (Подготовка мира)
	router, db := setupTestRouter("variant_search_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	// Создаем сложную структуру данных
	wh1 := h.CreateWarehouse("Склад 1")
	wh2 := h.CreateWarehouse("Склад 2")
	catShoes := h.CreateCategory("Обувь")
	catShirts := h.CreateCategory("Рубашки")
	unit := h.CreateUnit("пара")

	// Товар 1: Кроссовки Nike
	prodNike := h.CreateProduct(gin.H{"name": "Кроссовки Nike Air", "category_id": catShoes.ID})
	nikeRed42 := h.CreateVariant(gin.H{"product_id": prodNike.ID, "sku": "NK-AIR-RED-42", "unit_id": unit.ID})
	nikeBlack43 := h.CreateVariant(gin.H{"product_id": prodNike.ID, "sku": "NK-AIR-BLK-43", "unit_id": unit.ID})

	// Товар 2: Рубашка Поло
	prodPolo := h.CreateProduct(gin.H{"name": "Рубашка Поло", "category_id": catShirts.ID})
	poloBlueL := h.CreateVariant(gin.H{"product_id": prodPolo.ID, "sku": "POLO-BLUE-L", "unit_id": unit.ID})

	// Заполняем остатки
	// Склад 1: 10 красных кроссовок, 0 черных
	h.PostDocument(h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &wh1.ID, Items: []models.DocumentItem{
			{VariantID: nikeRed42.ID, Quantity: decimal.NewFromInt(10), Price: decimalPtr(decimal.Zero)},
		},
	}).ID)
	// Склад 2: 5 черных кроссовок, 20 синих рубашек
	h.PostDocument(h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &wh2.ID, Items: []models.DocumentItem{
			{VariantID: nikeBlack43.ID, Quantity: decimal.NewFromInt(5), Price: decimalPtr(decimal.Zero)},
			{VariantID: poloBlueL.ID, Quantity: decimal.NewFromInt(20), Price: decimalPtr(decimal.Zero)},
		},
	}).ID)

	// 2. ACT & ASSERT (Тестируем фильтры)
	t.Log("--- Проверка базовых фильтров ---")
	// Поиск по части имени "Nike" (должен найти 2 варианта)
	results := h.SearchVariants("name=Nike")
	h.Assert.Len(results, 2)
	h.Assert.Contains(variantIDs(results), nikeRed42.ID)
	h.Assert.Contains(variantIDs(results), nikeBlack43.ID)

	// Поиск по SKU (должен найти 1 вариант)
	results = h.SearchVariants("sku=POLO-BLUE-L")
	h.Assert.Len(results, 1)
	h.Assert.Equal(poloBlueL.ID, results[0].ID)

	// Поиск по категории "Рубашки" (должен найти 1 вариант)
	results = h.SearchVariants(fmt.Sprintf("category_id=%d", catShirts.ID))
	h.Assert.Len(results, 1)
	h.Assert.Equal(poloBlueL.ID, results[0].ID)

	t.Log("--- Проверка фильтров по наличию ---")
	// Ищем товары В НАЛИЧИИ на Складе 1 (только красные кроссовки)
	results = h.SearchVariants(fmt.Sprintf("stock_status=in_stock&warehouse_id=%d", wh1.ID))
	h.Assert.Len(results, 1)
	h.Assert.Equal(nikeRed42.ID, results[0].ID)

	// Ищем товары НЕ В НАЛИЧИИ на Складе 1 (черные кроссовки и синяя рубашка)
	results = h.SearchVariants(fmt.Sprintf("stock_status=out_of_stock&warehouse_id=%d", wh1.ID))
	h.Assert.Len(results, 2)
	h.Assert.Contains(variantIDs(results), nikeBlack43.ID)
	h.Assert.Contains(variantIDs(results), poloBlueL.ID)

	t.Log("--- Проверка комбинированных фильтров ---")
	// Ищем "Кроссовки" (по имени "Nike"), которые есть В НАЛИЧИИ на Складе 2
	queryParams := fmt.Sprintf("name=Nike&stock_status=in_stock&warehouse_id=%d", wh2.ID)
	results = h.SearchVariants(queryParams)
	h.Assert.Len(results, 1)
	h.Assert.Equal(nikeBlack43.ID, results[0].ID)

	t.Log("--- Проверка пагинации ---")
	// Запрашиваем все варианты (их 3), но с лимитом 1
	results = h.SearchVariants("limit=1")
	h.Assert.Len(results, 1)

	// Запрашиваем вторую страницу с лимитом 1 (offset=1)
	results = h.SearchVariants("limit=1&offset=1")
	h.Assert.Len(results, 1)
	// Проверяем, что это второй товар (по порядку ID)
	h.Assert.Equal(nikeBlack43.ID, results[0].ID)
}
