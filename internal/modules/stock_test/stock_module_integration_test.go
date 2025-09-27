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

func setupTestRouter() (*gin.Engine, *gorm.DB) {
	dsn := "file:lifecycle_test_db?mode=memory&cache=shared"
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

func TestDocumentLifecycle_Integration(t *testing.T) {
	router, db := setupTestRouter()
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

	// 2. --- ПОДГОТОВКА СПРАВОЧНИКОВ ---

	// Создаем Склад
	w := performRequest("POST", "/api/v1/stock/warehouses", gin.H{"name": "Основной склад"})
	assert.Equal(http.StatusCreated, w.Code)
	var createdWarehouse models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &createdWarehouse)
	assert.Equal("Основной склад", createdWarehouse.Name)

	// Создаем Категорию
	w = performRequest("POST", "/api/v1/stock/categories", gin.H{"name": "Ноутбуки"})
	assert.Equal(http.StatusCreated, w.Code)
	var createdCategory models.Category
	json.Unmarshal(w.Body.Bytes(), &createdCategory)

	// Создаем Единицу измерения
	w = performRequest("POST", "/api/v1/stock/units", gin.H{"name": "шт"})
	assert.Equal(http.StatusCreated, w.Code)
	var createdUnit models.Unit
	json.Unmarshal(w.Body.Bytes(), &createdUnit)

	// Создаем Товар
	itemPayload := models.Item{
		Name:       "MacBook Pro 16",
		SKU:        "MBP16-M3",
		UnitID:     createdUnit.ID,
		CategoryID: createdCategory.ID,
		Price:      2500,
	}
	w = performRequest("POST", "/api/v1/stock/items", itemPayload)
	assert.Equal(http.StatusCreated, w.Code)
	var createdItem models.Item
	json.Unmarshal(w.Body.Bytes(), &createdItem)
	assert.Equal("MBP16-M3", createdItem.SKU)

	// 3. --- РАБОТА С ДОКУМЕНТОМ ---

	// Создаем документ "Приход"
	docPayload := models.Document{
		Type:        "INCOME",
		WarehouseID: &createdWarehouse.ID,
		Comment:     "Первая поставка",
		Items: []models.DocumentItem{
			{ItemID: createdItem.ID, Quantity: 10},
		},
	}
	w = performRequest("POST", "/api/v1/stock/documents", docPayload)
	assert.Equal(http.StatusCreated, w.Code)
	var createdDoc models.Document
	json.Unmarshal(w.Body.Bytes(), &createdDoc)
	assert.Equal("draft", createdDoc.Status)

	// 4. --- ПРОВЕРКА ОСТАТКОВ (ДО ПРОВЕДЕНИЯ) ---
	balancePath := fmt.Sprintf("/api/v1/stock/balances/warehouse/%d", createdWarehouse.ID)
	w = performRequest("GET", balancePath, nil)
	assert.Equal(http.StatusOK, w.Code)
	var balancesBefore []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balancesBefore)
	assert.Empty(balancesBefore) // Остатков еще быть не должно

	// 5. --- ПРОВЕДЕНИЕ ДОКУМЕНТА ---
	postPath := fmt.Sprintf("/api/v1/stock/documents/%d/post", createdDoc.ID)
	w = performRequest("POST", postPath, nil)
	assert.Equal(http.StatusOK, w.Code)

	// 6. --- ПРОВЕРКА ОСТАТКОВ (ПОСЛЕ ПРОВЕДЕНИЯ) ---
	w = performRequest("GET", balancePath, nil)
	assert.Equal(http.StatusOK, w.Code)
	var balancesAfterPost []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balancesAfterPost)
	assert.Len(balancesAfterPost, 1) // Должна появиться одна запись об остатке
	assert.Equal(createdItem.ID, balancesAfterPost[0].ItemID)
	assert.Equal(10, balancesAfterPost[0].Quantity) // Проверяем количество

	// 7. --- ОТМЕНА ДОКУМЕНТА ---
	cancelPath := fmt.Sprintf("/api/v1/stock/documents/%d/cancel", createdDoc.ID)
	w = performRequest("POST", cancelPath, nil)
	assert.Equal(http.StatusOK, w.Code)

	// 8. --- ПРОВЕРКА ОСТАТКОВ (ПОСЛЕ ОТМЕНЫ) ---
	w = performRequest("GET", balancePath, nil)
	assert.Equal(http.StatusOK, w.Code)
	var balancesAfterCancel []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balancesAfterCancel)
	assert.Len(balancesAfterCancel, 1)
	assert.Equal(0, balancesAfterCancel[0].Quantity) // Проверяем, что количество списалось в ноль

	// Проверим и статус самого документа
	docPath := fmt.Sprintf("/api/v1/stock/documents/%d", createdDoc.ID)
	w = performRequest("GET", docPath, nil)
	assert.Equal(http.StatusOK, w.Code)
	var finalDoc models.Document
	json.Unmarshal(w.Body.Bytes(), &finalDoc)
	assert.Equal("canceled", finalDoc.Status)
}
