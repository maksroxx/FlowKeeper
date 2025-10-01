package stocktest

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestFifoQuantity_Integration(t *testing.T) {
	// Arrange
	router, db := setupTestRouter("fifo_test_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	t.Log("ПРИМЕЧАНИЕ: Этот тест ожидает, что в config/stock.yml установлена 'accounting_policy: fifo'")
	warehouse := h.CreateWarehouse("FIFO Склад")
	variant := h.CreateVariant(gin.H{"product_id": 1, "sku": "FIFO-TEST", "unit_id": 1})

	// 1. ПЕРВАЯ ПОСТАВКА: 10 шт.
	t.Log("Шаг 1: Первая поставка (10 шт)")
	incomeDoc1 := h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID, CreatedAt: time.Now(),
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(10), Price: decimalPtr(decimal.Zero)}},
	})
	h.PostDocument(incomeDoc1.ID)

	// 2. ВТОРАЯ ПОСТАВКА: 20 шт.
	t.Log("Шаг 2: Вторая поставка (20 шт)")
	time.Sleep(10 * time.Millisecond)
	incomeDoc2 := h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID, CreatedAt: time.Now(),
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(20), Price: decimalPtr(decimal.Zero)}},
	})
	h.PostDocument(incomeDoc2.ID)

	balance1 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(30).Equal(balance1.Quantity))

	// 3. ПРОДАЖА 15 штук
	t.Log("Шаг 3: Продажа 15 штук")
	h.PostDocument(h.CreateDocument(models.Document{
		Type: "OUTCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(15)}},
	}).ID)

	// Assert
	// 4. Проверяем финальный общий остаток
	finalBalance := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(15).Equal(finalBalance.Quantity))

	// 5. САМАЯ ГЛАВНАЯ ПРОВЕРКА FIFO:
	// Мы продали 15 шт. Первая партия (10 шт) должна быть полностью списана и УДАЛЕНА.
	// Во второй партии (20 шт) должно остаться 5 шт.

	var lots []models.StockLot
	db.Where("variant_id = ? AND warehouse_id = ? AND current_quantity > 0", variant.ID, warehouse.ID).Find(&lots)

	// ИСПРАВЛЕНО: Ожидаем ТОЛЬКО ОДНУ запись о партии с ненулевым остатком
	h.Assert.Len(lots, 1, "В БД должна остаться только одна партия с ненулевым остатком")

	// Проверяем, что это действительно остаток ВТОРОЙ партии
	remainingLot := lots[0]
	h.Assert.Equal(incomeDoc2.ID, remainingLot.IncomeDocumentID, "Оставшаяся партия должна быть из второго прихода")

	// Проверяем, что в ней правильное количество
	h.Assert.True(decimal.NewFromInt(15).Equal(remainingLot.CurrentQuantity), "Во второй партии должно было остаться 15 шт")

	t.Log("Логика FIFO списания и удаления пустых партий работает корректно.")
}
