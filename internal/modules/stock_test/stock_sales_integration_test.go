package stocktest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/maksroxx/flowkeeper/internal/modules/stock"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func setupTestRouterForSales() (*gin.Engine, *gorm.DB) {
	dsn := "file:sales_test_db?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	stockModule := stock.NewModule()

	err = stockModule.Migrate(db)
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate database: %v", err))
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	stockModule.RegisterRoutes(router, db)

	return router, db
}

func TestSalesLifecycle_Integration(t *testing.T) {
	router, db := setupTestRouterForSales()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	assert := require.New(t)

	performRequest := func(method, path string, body interface{}) *httptest.ResponseRecorder {
		var reqBody []byte
		if body != nil {
			reqBody, _ = json.Marshal(body)
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, path, bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		return w
	}

	w := performRequest("POST", "/api/v1/stock/warehouses", gin.H{"name": "Торговый зал"})
	assert.Equal(http.StatusCreated, w.Code)
	var warehouse models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &warehouse)

	w = performRequest("POST", "/api/v1/stock/units", gin.H{"name": "шт"})
	assert.Equal(http.StatusCreated, w.Code)
	var unit models.Unit
	json.Unmarshal(w.Body.Bytes(), &unit)

	w = performRequest("POST", "/api/v1/stock/items", gin.H{"name": "iPhone 15", "sku": "IP15", "unit_id": unit.ID})
	assert.Equal(http.StatusCreated, w.Code)
	var item models.Item
	json.Unmarshal(w.Body.Bytes(), &item)

	w = performRequest("POST", "/api/v1/stock/counterparties", gin.H{"name": "Иван Иванов"})
	assert.Equal(http.StatusCreated, w.Code)
	var customer models.Counterparty
	json.Unmarshal(w.Body.Bytes(), &customer)

	incomeDocPayload := models.Document{
		Type:        "INCOME",
		Number:      "TEST-IN-001",
		WarehouseID: &warehouse.ID,
		Items:       []models.DocumentItem{{ItemID: item.ID, Quantity: 15}},
	}
	w = performRequest("POST", "/api/v1/stock/documents", incomeDocPayload)
	assert.Equal(http.StatusCreated, w.Code)
	var incomeDoc models.Document
	json.Unmarshal(w.Body.Bytes(), &incomeDoc)

	w = performRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/post", incomeDoc.ID), nil)
	assert.Equal(http.StatusOK, w.Code)

	balancePath := fmt.Sprintf("/api/v1/stock/balances/warehouse/%d", warehouse.ID)
	w = performRequest("GET", balancePath, nil)
	var balancesAfterIncome []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balancesAfterIncome)
	assert.Len(balancesAfterIncome, 1)
	assert.Equal(15, balancesAfterIncome[0].Quantity)

	saleDocPayload := models.Document{
		Type:           "OUTCOME",
		Number:         "TEST-SALE-001",
		WarehouseID:    &warehouse.ID,
		CounterpartyID: &customer.ID,
		Items:          []models.DocumentItem{{ItemID: item.ID, Quantity: 5}},
		Comment:        "Продажа 5 айфонов",
	}
	w = performRequest("POST", "/api/v1/stock/documents", saleDocPayload)
	assert.Equal(http.StatusCreated, w.Code)
	var saleDoc models.Document
	json.Unmarshal(w.Body.Bytes(), &saleDoc)

	w = performRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/post", saleDoc.ID), nil)
	assert.Equal(http.StatusOK, w.Code)

	w = performRequest("GET", balancePath, nil)
	var balancesAfterSale []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balancesAfterSale)
	assert.Len(balancesAfterSale, 1)
	assert.Equal(10, balancesAfterSale[0].Quantity, "Остаток на складе должен был уменьшиться до 10")

	w = performRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/cancel", saleDoc.ID), nil)
	assert.Equal(http.StatusOK, w.Code)

	w = performRequest("GET", balancePath, nil)
	var balancesAfterCancel []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balancesAfterCancel)
	assert.Len(balancesAfterCancel, 1)
	assert.Equal(15, balancesAfterCancel[0].Quantity, "Остаток должен был вернуться к 15 после отмены")
}
