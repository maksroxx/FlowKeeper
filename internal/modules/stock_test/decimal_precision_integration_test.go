package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestDecimalPrecision_Integration(t *testing.T) {
	router, db := setupTestRouter("decimal_precision_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	warehouse := h.CreateWarehouse("Склад весовых товаров")
	variant := h.CreateVariant(gin.H{"product_id": 1, "sku": "WEIGHT-TEST"})

	incomeQty := decimal.NewFromInt(10)
	incomePrice, _ := decimal.NewFromString("3.33")

	incomeDoc := h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: incomeQty, Price: decimalPtr(incomePrice)}},
	})
	h.PostDocument(incomeDoc.ID)

	outcomeQty := decimal.NewFromInt(10).DivRound(decimal.NewFromInt(3), 4)

	for i := 0; i < 3; i++ {
		h.PostDocument(h.CreateDocument(models.Document{
			Type: "OUTCOME", WarehouseID: &warehouse.ID,
			Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: outcomeQty}},
		}).ID)
	}

	finalBalance := h.GetBalances(warehouse.ID)[0]

	expectedFinalQty, _ := decimal.NewFromString("0.0001")

	h.Assert.True(expectedFinalQty.Equal(finalBalance.Quantity),
		"Финальный остаток количества должен быть %s, а он: %s", expectedFinalQty.String(), finalBalance.Quantity.String())
}
