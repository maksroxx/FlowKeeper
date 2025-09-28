package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestInventoryDocument_CostRecalculation(t *testing.T) {
	// Arrange
	router, db := setupTestRouter("inventory_cost_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	warehouse := h.CreateWarehouse("Склад для инвентаризации")
	variant := h.CreateVariant(gin.H{"product_id": 1, "sku": "INV-TEST"})

	// 1. Приходуем 10 шт по 100
	initialQty := decimal.NewFromInt(10)
	initialPrice := decimal.NewFromInt(100)
	h.PostDocument(h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: initialQty, Price: decimalPtr(initialPrice)}},
	}).ID)

	// Проверяем исходное состояние
	balance1 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(initialQty.Equal(balance1.Quantity))
	h.Assert.True(decimal.NewFromInt(1000).Equal(balance1.TotalCost)) // 10 * 100 = 1000

	// 2. Act: Проводим инвентаризацию. Обнаружили "излишек" в 2 шт.
	// Мы указываем ФАКТИЧЕСКОЕ количество на складе.
	t.Log("Проведение инвентаризации (обнаружен излишек)")
	inventoryQty := decimal.NewFromInt(12)
	inventoryDoc := h.CreateDocument(models.Document{
		Type: "INVENTORY", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: inventoryQty, Price: decimalPtr(decimal.Zero)}}, // Цена для инвентаризации обычно 0
	})
	h.PostDocument(inventoryDoc.ID)

	// 3. Assert: Проверяем пересчет остатков и себестоимости
	balance2 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(inventoryQty.Equal(balance2.Quantity), "Количество после инвентаризации должно стать 12")
	// Общая стоимость не изменилась, т.к. излишек пришел с нулевой ценой
	h.Assert.True(decimal.NewFromInt(1000).Equal(balance2.TotalCost), "Общая стоимость не должна была измениться")

	// 4. Act: Продаем 1 шт товара
	t.Log("Продажа после инвентаризации для проверки новой себестоимости")
	h.PostDocument(h.CreateDocument(models.Document{
		Type: "OUTCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(1)}},
	}).ID)

	// 5. Assert: Проверяем, что списание произошло по НОВОЙ средней себестоимости
	finalBalance := h.GetBalances(warehouse.ID)[0]
	// Новая средняя себестоимость: 1000 / 12 = 83.3333
	newAverageCost := decimal.NewFromInt(1000).DivRound(decimal.NewFromInt(12), 4)

	// Ожидаемый финальный остаток: 12 - 1 = 11
	expectedFinalQty := decimal.NewFromInt(11)
	// Ожидаемая финальная стоимость: 1000 - (1 * 83.3333) = 916.6667
	expectedFinalCost := decimal.NewFromInt(1000).Sub(newAverageCost)

	h.Assert.True(expectedFinalQty.Equal(finalBalance.Quantity), "Финальное количество")
	h.Assert.True(expectedFinalCost.Equal(finalBalance.TotalCost), "Финальная себестоимость должна была списаться по новой средней")
	t.Logf("Новая средняя себестоимость: %s", newAverageCost.String())
	t.Logf("Финальная общая стоимость: %s", finalBalance.TotalCost.String())
}
