package stocktest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/maksroxx/flowkeeper/internal/modules/stock"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

func setupTestRouter(dbName string) (*gin.Engine, *gorm.DB) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
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

type TestHelper struct {
	T      *testing.T
	Router *gin.Engine
	Assert *require.Assertions
}

func NewTestHelper(t *testing.T, router *gin.Engine) *TestHelper {
	return &TestHelper{
		T:      t,
		Router: router,
		Assert: require.New(t),
	}
}

// PerformRequest - низкоуровневый метод для выполнения любого HTTP-запроса к тестовому серверу.
func (h *TestHelper) PerformRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	h.Router.ServeHTTP(w, req)
	return w
}

// CreateDocument создает новый документ через API (POST /documents).
func (h *TestHelper) CreateDocument(payload interface{}) models.Document {
	w := h.PerformRequest("POST", "/api/v1/stock/documents", payload)
	h.Assert.Equal(http.StatusCreated, w.Code, fmt.Sprintf("Failed to create document. Body: %s", w.Body.String()))
	var created models.Document
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// GetDocument получает один документ по его ID через API (GET /documents/:id).
func (h *TestHelper) GetDocument(docID uint) models.Document {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/documents/%d", docID), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var doc models.Document
	json.Unmarshal(w.Body.Bytes(), &doc)
	return doc
}

// ListDocuments получает список документов через API (GET /documents), опционально фильтруя по статусу.
func (h *TestHelper) ListDocuments(status ...string) []models.Document {
	path := "/api/v1/stock/documents"
	if len(status) > 0 && status[0] != "" {
		path = fmt.Sprintf("%s?status=%s", path, status[0])
	}
	w := h.PerformRequest("GET", path, nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var docs []models.Document
	json.Unmarshal(w.Body.Bytes(), &docs)
	return docs
}

// UpdateDocument обновляет существующий документ через API (PUT /documents/:id).
func (h *TestHelper) UpdateDocument(docID uint, payload interface{}) models.Document {
	w := h.PerformRequest("PUT", fmt.Sprintf("/api/v1/stock/documents/%d", docID), payload)
	h.Assert.Equal(http.StatusOK, w.Code)
	var doc models.Document
	json.Unmarshal(w.Body.Bytes(), &doc)
	return doc
}

// DeleteDocument удаляет документ по его ID через API (DELETE /documents/:id).
func (h *TestHelper) DeleteDocument(docID uint) {
	w := h.PerformRequest("DELETE", fmt.Sprintf("/api/v1/stock/documents/%d", docID), nil)
	h.Assert.Equal(http.StatusNoContent, w.Code)
}

// PostDocument проводит документ по его ID через API (POST /documents/:id/post).
func (h *TestHelper) PostDocument(docID uint) {
	w := h.PerformRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/post", docID), nil)
	h.Assert.Equal(http.StatusOK, w.Code, fmt.Sprintf("Failed to post document. Body: %s", w.Body.String()))
}

// CancelDocument отменяет проведенный документ по ID через API (POST /documents/:id/cancel).
func (h *TestHelper) CancelDocument(docID uint) {
	w := h.PerformRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/cancel", docID), nil)
	h.Assert.Equal(http.StatusOK, w.Code, fmt.Sprintf("Failed to cancel document. Body: %s", w.Body.String()))
}

// CreateItem создает новый товар (номенклатуру) через API (POST /items).
func (h *TestHelper) CreateItem(payload gin.H) models.Item {
	w := h.PerformRequest("POST", "/api/v1/stock/items", payload)
	h.Assert.Equal(http.StatusCreated, w.Code, "Failed to create item")
	var created models.Item
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// GetItem получает один товар по его ID через API (GET /items/:id).
func (h *TestHelper) GetItem(itemID uint) models.Item {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/items/%d", itemID), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var item models.Item
	json.Unmarshal(w.Body.Bytes(), &item)
	return item
}

// ListItems получает список всех товаров через API (GET /items).
func (h *TestHelper) ListItems() []models.Item {
	w := h.PerformRequest("GET", "/api/v1/stock/items", nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var items []models.Item
	json.Unmarshal(w.Body.Bytes(), &items)
	return items
}

// CreateWarehouse создает новый склад через API (POST /warehouses).
func (h *TestHelper) CreateWarehouse(name string) models.Warehouse {
	w := h.PerformRequest("POST", "/api/v1/stock/warehouses", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code, "Failed to create warehouse")
	var created models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// GetWarehouse получает один склад по его ID через API (GET /warehouses/:id).
func (h *TestHelper) GetWarehouse(id uint) models.Warehouse {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/warehouses/%d", id), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var wh models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &wh)
	return wh
}

// ListWarehouses получает список всех складов через API (GET /warehouses).
func (h *TestHelper) ListWarehouses() []models.Warehouse {
	w := h.PerformRequest("GET", "/api/v1/stock/warehouses", nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var whs []models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &whs)
	return whs
}

// CreatePriceType создает новый тип цены (например, "Розничная") через API (POST /price-types).
func (h *TestHelper) CreatePriceType(name string) models.PriceType {
	w := h.PerformRequest("POST", "/api/v1/stock/price-types", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code, "Failed to create price type")
	var created models.PriceType
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// GetPrice получает актуальную цену для товара и типа цены через API (GET /prices).
func (h *TestHelper) GetPrice(itemID, priceTypeID uint) models.ItemPrice {
	path := fmt.Sprintf("/api/v1/stock/prices?item_id=%d&price_type_id=%d", itemID, priceTypeID)
	w := h.PerformRequest("GET", path, nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var price models.ItemPrice
	json.Unmarshal(w.Body.Bytes(), &price)
	return price
}

// CreateCategory создает новую категорию товаров через API (POST /categories).
func (h *TestHelper) CreateCategory(name string) models.Category {
	w := h.PerformRequest("POST", "/api/v1/stock/categories", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code, "Failed to create category")
	var created models.Category
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// CreateUnit создает новую единицу измерения через API (POST /units).
func (h *TestHelper) CreateUnit(name string) models.Unit {
	w := h.PerformRequest("POST", "/api/v1/stock/units", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code, "Failed to create unit")
	var created models.Unit
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// CreateCounterparty создает нового контрагента (клиента/поставщика) через API (POST /counterparties).
func (h *TestHelper) CreateCounterparty(payload gin.H) models.Counterparty {
	w := h.PerformRequest("POST", "/api/v1/stock/counterparties", payload)
	h.Assert.Equal(http.StatusCreated, w.Code, fmt.Sprintf("Failed to create counterparty. Body: %s", w.Body.String()))
	var created models.Counterparty
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// GetCounterparty возвращает контрагента чурущ API (GET /conterpartirs/:id)
func (h *TestHelper) GetCounterparty(id uint) models.Counterparty {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/counterparties/%d", id), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var cp models.Counterparty
	json.Unmarshal(w.Body.Bytes(), &cp)
	return cp
}

// GetBalances получает все остатки на конкретном складе через API (GET /balances/warehouse/:id).
func (h *TestHelper) GetBalances(warehouseID uint) []models.StockBalance {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/balances/warehouse/%d", warehouseID), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var balances []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balances)
	return balances
}

// ListMovements получает список всех движений товаров через API (GET /movements).
func (h *TestHelper) ListMovements() []models.StockMovement {
	w := h.PerformRequest("GET", "/api/v1/stock/movements", nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var movements []models.StockMovement
	json.Unmarshal(w.Body.Bytes(), &movements)
	return movements
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
