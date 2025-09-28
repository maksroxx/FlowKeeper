package stocktest

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestBalanceFiltering_Integration(t *testing.T) {
	router, db := setupTestRouter("filtering_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	warehouse1 := h.CreateWarehouse("Склад 1")
	warehouse2 := h.CreateWarehouse("Склад 2")
	catLaptops := h.CreateCategory("Ноутбуки")
	catPhones := h.CreateCategory("Телефоны")
	unit := h.CreateUnit("шт")

	prodMacbook := h.CreateProduct(gin.H{"name": "MacBook", "category_id": catLaptops.ID, "unit_id": unit.ID})
	varMacbookPro := h.CreateVariant(gin.H{"product_id": prodMacbook.ID, "sku": "MBP-16", "unit_id": unit.ID})

	prodIphone := h.CreateProduct(gin.H{"name": "iPhone", "category_id": catPhones.ID, "unit_id": unit.ID})
	varIphone15 := h.CreateVariant(gin.H{"product_id": prodIphone.ID, "sku": "IP-15", "unit_id": unit.ID})

	h.PostDocument(h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse1.ID, Items: []models.DocumentItem{
			{VariantID: varMacbookPro.ID, Quantity: decimal.NewFromInt(10), Price: decimalPtr(decimal.Zero)},
			{VariantID: varIphone15.ID, Quantity: decimal.NewFromInt(5), Price: decimalPtr(decimal.Zero)},
		},
	}).ID)
	h.PostDocument(h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse2.ID, Items: []models.DocumentItem{
			{VariantID: varIphone15.ID, Quantity: decimal.NewFromInt(3), Price: decimalPtr(decimal.Zero)},
		},
	}).ID)

	t.Log("Тестирование фильтров для Склада 1")
	allBalances := h.GetBalancesFiltered(warehouse1.ID, "")
	h.Assert.Len(allBalances, 2)

	balanceSku := h.GetBalancesFiltered(warehouse1.ID, "sku=MBP-16")
	h.Assert.Len(balanceSku, 1)
	h.Assert.Equal(varMacbookPro.ID, balanceSku[0].VariantID)

	balanceSkuNotFound := h.GetBalancesFiltered(warehouse1.ID, "sku=NON-EXISTENT")
	h.Assert.Len(balanceSkuNotFound, 0)

	balanceCat := h.GetBalancesFiltered(warehouse1.ID, fmt.Sprintf("category_id=%d", catPhones.ID))
	h.Assert.Len(balanceCat, 1)
	h.Assert.Equal(varIphone15.ID, balanceCat[0].VariantID)

	balanceQty := h.GetBalancesFiltered(warehouse1.ID, "min_qty=8")
	h.Assert.Len(balanceQty, 1)
	h.Assert.Equal(varMacbookPro.ID, balanceQty[0].VariantID)

	balanceCombo := h.GetBalancesFiltered(warehouse1.ID, fmt.Sprintf("category_id=%d&sku=MBP-16", catLaptops.ID))
	h.Assert.Len(balanceCombo, 1)
	h.Assert.Equal(varMacbookPro.ID, balanceCombo[0].VariantID)

	balanceComboNone := h.GetBalancesFiltered(warehouse1.ID, fmt.Sprintf("category_id=%d&sku=IP-15", catLaptops.ID))
	h.Assert.Len(balanceComboNone, 0)

	t.Log("Тестирование фильтров для Склада 2")
	allBalancesW2 := h.GetBalancesFiltered(warehouse2.ID, "")
	h.Assert.Len(allBalancesW2, 1)
	h.Assert.Equal(varIphone15.ID, allBalancesW2[0].VariantID)
}
