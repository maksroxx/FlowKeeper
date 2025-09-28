package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/shopspring/decimal"
)

func TestBalances_Integration(t *testing.T) {
	router, db := setupTestRouter("balances_test_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	warehouse := h.CreateWarehouse("Главный склад")
	unit := h.CreateUnit("кг")
	category := h.CreateCategory("Фрукты")
	item := h.CreateItem(gin.H{"name": "Яблоки", "sku": "APL-GR", "unit_id": unit.ID, "category_id": category.ID})

	incomeQty := decimal.NewFromFloat(100.5)
	incomePrice := decimal.NewFromFloat(50.0)
	incomeTotalCost := incomeQty.Mul(incomePrice)

	outcomeQty := decimal.NewFromFloat(30.2)

	incomeDoc := h.CreateDocument(models.Document{
		Type:        "INCOME",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: incomeQty, Price: decimalPtr(incomePrice)}},
	})
	h.PostDocument(incomeDoc.ID)

	balancesAfterIncome := h.GetBalances(warehouse.ID)
	h.Assert.Len(balancesAfterIncome, 1)
	h.Assert.True(incomeQty.Equal(balancesAfterIncome[0].Quantity), "Остаток после прихода должен быть 100.5")
	h.Assert.True(incomeTotalCost.Equal(balancesAfterIncome[0].TotalCost), "Общая стоимость после прихода должна быть 5025.0")

	outcomeDoc := h.CreateDocument(models.Document{
		Type:        "OUTCOME",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: outcomeQty}},
	})
	h.PostDocument(outcomeDoc.ID)

	finalBalances := h.GetBalances(warehouse.ID)
	h.Assert.Len(finalBalances, 1, "Should be one balance record for apples")

	expectedFinalQty := incomeQty.Sub(outcomeQty)
	outcomeTotalCost := outcomeQty.Mul(incomePrice)
	expectedFinalCost := incomeTotalCost.Sub(outcomeTotalCost)

	h.Assert.True(expectedFinalQty.Equal(finalBalances[0].Quantity), "The final quantity should be 70.3")
	h.Assert.True(expectedFinalCost.Equal(finalBalances[0].TotalCost), "The final total cost should be 3515.0")
}
