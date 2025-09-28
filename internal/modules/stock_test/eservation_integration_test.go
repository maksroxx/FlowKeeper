package stocktest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func TestReservationLifecycle_Integration(t *testing.T) {
	router, db := setupTestRouter("reservation_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	h := NewTestHelper(t, router)

	warehouse := h.CreateWarehouse("Склад для Заказов")
	variant := h.CreateVariant(gin.H{"product_id": 1, "sku": "RES-TEST"})

	incomeDoc := h.CreateDocument(models.Document{
		Type: "INCOME", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(20), Price: decimalPtr(decimal.NewFromInt(1))}},
	})
	h.PostDocument(incomeDoc.ID)

	orderDoc := h.CreateDocument(models.Document{
		Type: "ORDER", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(15)}},
	})
	h.PostDocument(orderDoc.ID)

	balance1 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(20).Equal(balance1.Quantity), "Физический остаток не должен измениться")

	orderDoc2 := h.CreateDocument(models.Document{
		Type: "ORDER", WarehouseID: &warehouse.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(10)}},
	})
	w := h.PerformRequest("POST", "/api/v1/stock/documents/"+fmt.Sprintf("%d", orderDoc2.ID)+"/post", nil)
	h.Assert.Equal(http.StatusInternalServerError, w.Code, "Проведение заказа сверх доступного остатка должно вернуть ошибку")
	h.Assert.Contains(w.Body.String(), "not enough available stock")

	outcomeDoc := h.CreateDocument(models.Document{
		Type: "OUTCOME", WarehouseID: &warehouse.ID, BaseDocumentID: &orderDoc.ID,
		Items: []models.DocumentItem{{VariantID: variant.ID, Quantity: decimal.NewFromInt(15)}},
	})
	h.PostDocument(outcomeDoc.ID)

	balance2 := h.GetBalances(warehouse.ID)[0]
	h.Assert.True(decimal.NewFromInt(5).Equal(balance2.Quantity), "Физический остаток должен уменьшиться до 5")
}
