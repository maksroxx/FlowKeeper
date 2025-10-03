package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestInventoryLifecycle_Integration(t *testing.T) {
	// Arrange
	router, db := setupTestRouter("lifecycle_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	warehouse := h.CreateWarehouse("Склад Жизненного Цикла")
	variant := h.CreateVariant(gin.H{
		"product_id": h.CreateProduct(gin.H{"name": "Тестовый Товар"}).ID,
		"sku":        "LFC-TEST",
		"unit_id":    h.CreateUnit("шт").ID,
	})
	incomePrice := decimal.NewFromInt(100)

	// 1. Приход 10 штук по 100
	incomeDoc := h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(10), Price: decimalPtr(incomePrice)}},
	})
	h.PostDocument(incomeDoc.ID)

	balance1 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(10).Equal(balance1.Quantity))

	// 2. Расход 3 штук
	outcomeDoc := h.CreateDocument(models.Document{
		Type: "OUTCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(3)}},
	})
	h.PostDocument(outcomeDoc.ID)

	balance2 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(7).Equal(balance2.Quantity)) // 10 - 3 = 7

	// 3. Отмена расхода (товар должен вернуться)
	h.CancelDocument(outcomeDoc.ID)

	balance3 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(10).Equal(balance3.Quantity), "Остаток должен вернуться к 10")

	// 4. Отмена прихода (остаток должен стать нулевым)
	h.CancelDocument(incomeDoc.ID)

	finalBalances := h.GetBalances(warehouse.ID)
	h.Assert.Len(finalBalances, 1)
	h.Assert.True(decimal.Zero.Equal(finalBalances[0].Quantity), "Остаток должен стать 0")
}
