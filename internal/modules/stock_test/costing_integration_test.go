package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

// Первая поставка: Приходуем 10 штук товара по 100 рублей. (Средняя себестоимость = 100).
// Продажа: Продаем 4 штуки. Система должна списать их по себестоимости 100 рублей. (Остаток: 6 шт, общая стоимость 600).
// Вторая поставка (дороже): Приходуем еще 10 штук, но уже по 130 рублей. (Остаток: 16 шт, общая стоимость 600 + 1300 = 1900). Средняя себестоимость теперь изменилась: 1900 / 16 = 118.75.
// Вторая продажа: Продаем 2 штуки. Система должна списать их уже по новой средней себестоимости 118.75.
// Проверяем финальные остатки: Количество и общая стоимость должны сходиться с нашими ручными расчётами.

func TestCostCalculation_Integration(t *testing.T) {
	router, db := setupTestRouter("costing_test_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	warehouse := h.CreateWarehouse("Основной склад")
	item := h.CreateItem(gin.H{
		"name":        "Товар для теста себестоимости",
		"sku":         "COST-TEST",
		"unit_id":     h.CreateUnit("шт").ID,
		"category_id": h.CreateCategory("Тест").ID,
	})

	// --- 1. ПЕРВАЯ ПОСТАВКА: 10 шт по 100 ---
	t.Log("Шаг 1: Первая поставка (10 шт по 100)")
	price1 := decimal.NewFromInt(100)
	qty1 := decimal.NewFromInt(10)
	incomeDoc1 := h.CreateDocument(models.Document{
		Type:        "INCOME",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: qty1, Price: decimalPtr(price1)}},
	})
	h.PostDocument(incomeDoc1.ID)

	balance1 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(qty1.Equal(balance1.Quantity), "Остаток после 1-го прихода (кол-во)")
	h.Assert.True(decimal.NewFromInt(1000).Equal(balance1.TotalCost), "Остаток после 1-го прихода (сумма)") // 10 * 100 = 1000

	// --- 2. ПЕРВАЯ ПРОДАЖА: 4 шт (списываются по 100) ---
	t.Log("Шаг 2: Первая продажа (4 шт)")
	qty_sale1 := decimal.NewFromInt(4)
	outcomeDoc1 := h.CreateDocument(models.Document{
		Type:        "OUTCOME",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: qty_sale1}},
	})
	h.PostDocument(outcomeDoc1.ID)

	// Проверяем остаток
	balance2 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(6).Equal(balance2.Quantity), "Остаток после 1-й продажи (кол-во)")   // 10 - 4 = 6
	h.Assert.True(decimal.NewFromInt(600).Equal(balance2.TotalCost), "Остаток после 1-й продажи (сумма)") // 6 * 100 = 600

	// --- 3. ВТОРАЯ ПОСТАВКА: 10 шт по 130 (дороже) ---
	t.Log("Шаг 3: Вторая поставка (10 шт по 130)")
	price2 := decimal.NewFromInt(130)
	qty2 := decimal.NewFromInt(10)
	incomeDoc2 := h.CreateDocument(models.Document{
		Type:        "INCOME",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: qty2, Price: decimalPtr(price2)}},
	})
	h.PostDocument(incomeDoc2.ID)

	// Проверяем остаток. Теперь средняя себестоимость изменилась.
	balance3 := h.GetBalances(warehouse.ID)[0]
	// Кол-во: 6 + 10 = 16
	h.Assert.True(decimal.NewFromInt(16).Equal(balance3.Quantity), "Остаток после 2-го прихода (кол-во)")
	// Сумма: 600 (старые) + 1300 (новые) = 1900
	h.Assert.True(decimal.NewFromInt(1900).Equal(balance3.TotalCost), "Остаток после 2-го прихода (сумма)")

	// --- 4. ВТОРАЯ ПРОДАЖА: 2 шт (должны списаться по новой средней себестоимости) ---
	t.Log("Шаг 4: Вторая продажа (2 шт)")
	qty_sale2 := decimal.NewFromInt(2)
	// Рассчитанная средняя себестоимость: 1900 / 16 = 118.75
	expectedAverageCost := decimal.NewFromFloat(118.75)

	outcomeDoc2 := h.CreateDocument(models.Document{
		Type:        "OUTCOME",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: qty_sale2}},
	})
	h.PostDocument(outcomeDoc2.ID)

	// --- 5. ФИНАЛЬНАЯ ПРОВЕРКА ---
	t.Log("Шаг 5: Финальная проверка остатков")
	finalBalance := h.GetBalances(warehouse.ID)[0]

	// Ожидаемое финальное количество: 16 - 2 = 14
	expectedFinalQty := decimal.NewFromInt(14)
	h.Assert.True(expectedFinalQty.Equal(finalBalance.Quantity), "Финальный остаток (кол-во)")

	// Ожидаемая финальная сумма: 1900 - (2 * 118.75) = 1900 - 237.5 = 1662.5
	costOfSale2 := qty_sale2.Mul(expectedAverageCost)              // 2 * 118.75 = 237.5
	expectedFinalCost := decimal.NewFromInt(1900).Sub(costOfSale2) // 1900 - 237.5 = 1662.5

	h.Assert.True(expectedFinalCost.Equal(finalBalance.TotalCost), "Финальный остаток (сумма)")
	t.Logf("Финальная себестоимость: %s, ожидалось: %s", finalBalance.TotalCost.String(), expectedFinalCost.String())
}
