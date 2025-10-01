package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestFullE2E_BusinessCycle(t *testing.T) {
	router, db := setupTestRouter("full_e2e_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	t.Log("Шаг 1: Создание справочников")
	warehouse := h.CreateWarehouse("Основной склад")
	category := h.CreateCategory("Одежда")
	unit := h.CreateUnit("шт")
	priceTypeRetail := h.CreatePriceType("Розничная")
	customer := h.CreateCounterparty(gin.H{"name": "Розничный покупатель"})

	t.Log("Шаг 2: Создание номенклатуры и ее вариантов")
	productTShirt := h.CreateProduct(gin.H{"name": "Футболка Поло", "category_id": category.ID})

	variantBlue := h.CreateVariant(gin.H{
		"product_id": productTShirt.ID, "sku": "POLO-BLUE-L", "unit_id": unit.ID,
		"characteristics": gin.H{"Цвет": "Синий", "Размер": "L"},
	})
	variantRed := h.CreateVariant(gin.H{
		"product_id": productTShirt.ID, "sku": "POLO-RED-M", "unit_id": unit.ID,
		"characteristics": gin.H{"Цвет": "Красный", "Размер": "M"},
	})

	t.Log("Шаг 3: Установка цен")
	priceBlue := decimal.NewFromInt(1500)
	priceRed := decimal.NewFromInt(1600)
	priceDoc := h.CreateDocument(models.Document{
		Type:        "PRICE_UPDATE",
		PriceTypeID: &priceTypeRetail.ID,
		Items: []models.DocumentItem{
			{VariantID: variantBlue.ID, Price: decimalPtr(priceBlue)},
			{VariantID: variantRed.ID, Price: decimalPtr(priceRed)},
		},
	})
	h.PostDocument(priceDoc.ID)
	h.Assert.True(priceBlue.Equal(h.GetPrice(variantBlue.ID, priceTypeRetail.ID).Price))

	t.Log("Шаг 4: Поступление товаров на склад")
	costBlue := decimal.NewFromInt(800)
	costRed := decimal.NewFromInt(850)
	incomeDoc := h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{
			{VariantID: variantBlue.ID, Quantity: decimal.NewFromInt(20), Price: decimalPtr(costBlue)},
			{VariantID: variantRed.ID, Quantity: decimal.NewFromInt(15), Price: decimalPtr(costRed)},
		},
	})
	h.PostDocument(incomeDoc.ID)
	balances1 := h.GetBalances(warehouse.ID)
	h.Assert.Len(balances1, 2)
	h.Assert.True(decimal.NewFromInt(20).Equal(findBalance(balances1, variantBlue.ID).Quantity))

	t.Log("Шаг 5: Создание и проведение заказа клиента")
	orderDoc := h.CreateDocument(models.Document{
		Type:           "ORDER",
		WarehouseID:    &warehouse.ID,
		CounterpartyID: &customer.ID,
		Items: []models.DocumentItem{
			{VariantID: variantBlue.ID, Quantity: decimal.NewFromInt(5)},
			{VariantID: variantRed.ID, Quantity: decimal.NewFromInt(2)},
		},
	})
	h.PostDocument(orderDoc.ID)

	balances2 := h.GetBalances(warehouse.ID)
	h.Assert.True(decimal.NewFromInt(20).Equal(findBalance(balances2, variantBlue.ID).Quantity), "Физический остаток синих не должен измениться")

	t.Log("Шаг 6: Продажа на основании заказа")
	outcomeDoc := h.CreateDocument(models.Document{
		Type:           "OUTCOME",
		BaseDocumentID: &orderDoc.ID,
		WarehouseID:    &warehouse.ID,
		CounterpartyID: &customer.ID,
		Items:          orderDoc.Items,
	})
	h.PostDocument(outcomeDoc.ID)

	t.Log("Шаг 7: Финальная проверка остатков и движений")
	finalBalances := h.GetBalances(warehouse.ID)

	finalBalanceBlue := findBalance(finalBalances, variantBlue.ID)
	h.Assert.True(decimal.NewFromInt(15).Equal(finalBalanceBlue.Quantity), "Финальный остаток синих футболок")

	finalBalanceRed := findBalance(finalBalances, variantRed.ID)
	h.Assert.True(decimal.NewFromInt(13).Equal(finalBalanceRed.Quantity), "Финальный остаток красных футболок")

	movements := h.ListMovements()
	h.Assert.NotEmpty(movements)

	finalOrderDoc := h.GetDocument(orderDoc.ID)
	finalOutcomeDoc := h.GetDocument(outcomeDoc.ID)
	h.Assert.Equal("posted", finalOrderDoc.Status)
	h.Assert.Equal("posted", finalOutcomeDoc.Status)
	h.Assert.Equal(orderDoc.ID, *finalOutcomeDoc.BaseDocumentID)
}
