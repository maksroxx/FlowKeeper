package stocktest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	dbPath := fmt.Sprintf("test_%s.db", dbName)
	os.Remove(dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	db.Exec("PRAGMA journal_mode = WAL;")

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

// TestHelper - наш помощник для интеграционных тестов.
type TestHelper struct {
	T      *testing.T
	Router *gin.Engine
	Assert *require.Assertions
}

func NewTestHelper(t *testing.T, router *gin.Engine) *TestHelper {
	return &TestHelper{T: t, Router: router, Assert: require.New(t)}
}

// PerformRequest - низкоуровневый метод для выполнения любого HTTP-запроса.
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

// --- Утилиты ---
func decimalPtr(d decimal.Decimal) *decimal.Decimal { return &d }
func findBalance(balances []models.StockBalance, variantID uint) models.StockBalance {
	for _, b := range balances {
		if b.VariantID == variantID {
			return b
		}
	}
	return models.StockBalance{}
}

// --- Products ---
func (h *TestHelper) CreateProduct(payload gin.H) models.Product {
	w := h.PerformRequest("POST", "/api/v1/stock/products", payload)
	h.Assert.Equal(http.StatusCreated, w.Code, fmt.Sprintf("Failed to create product. Body: %s", w.Body.String()))
	var created models.Product
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}
func (h *TestHelper) GetProduct(id uint) models.Product {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/products/%d", id), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var p models.Product
	json.Unmarshal(w.Body.Bytes(), &p)
	return p
}
func (h *TestHelper) ListProducts() []models.Product {
	w := h.PerformRequest("GET", "/api/v1/stock/products", nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var ps []models.Product
	json.Unmarshal(w.Body.Bytes(), &ps)
	return ps
}
func (h *TestHelper) UpdateProduct(id uint, payload gin.H) models.Product {
	w := h.PerformRequest("PUT", fmt.Sprintf("/api/v1/stock/products/%d", id), payload)
	h.Assert.Equal(http.StatusOK, w.Code)
	var updated models.Product
	json.Unmarshal(w.Body.Bytes(), &updated)
	return updated
}
func (h *TestHelper) DeleteProduct(id uint) {
	w := h.PerformRequest("DELETE", fmt.Sprintf("/api/v1/stock/products/%d", id), nil)
	h.Assert.Equal(http.StatusNoContent, w.Code)
}

// --- Variants ---
func (h *TestHelper) CreateVariant(payload gin.H) models.Variant {
	w := h.PerformRequest("POST", "/api/v1/stock/variants", payload)
	h.Assert.Equal(http.StatusCreated, w.Code, fmt.Sprintf("Failed to create variant. Body: %s", w.Body.String()))
	var created models.Variant
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}
func (h *TestHelper) GetVariant(id uint) models.Variant {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/variants/%d", id), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var v models.Variant
	json.Unmarshal(w.Body.Bytes(), &v)
	return v
}

//	func (h *TestHelper) ListVariants() []models.Variant {
//		w := h.PerformRequest("GET", "/api/v1/stock/variants", nil)
//		h.Assert.Equal(http.StatusOK, w.Code)
//		var vs []models.Variant
//		json.Unmarshal(w.Body.Bytes(), &vs)
//		return vs
//	}

func (h *TestHelper) SearchVariants(queryParams string) []models.Variant {
	path := "/api/v1/stock/variants"
	if queryParams != "" {
		path = fmt.Sprintf("%s?%s", path, queryParams)
	}
	w := h.PerformRequest("GET", path, nil)
	h.Assert.Equal(http.StatusOK, w.Code, fmt.Sprintf("Failed to search variants. Body: %s", w.Body.String()))
	var variants []models.Variant
	json.Unmarshal(w.Body.Bytes(), &variants)
	return variants
}

func (h *TestHelper) UpdateVariant(id uint, payload gin.H) models.Variant {
	w := h.PerformRequest("PUT", fmt.Sprintf("/api/v1/stock/variants/%d", id), payload)
	h.Assert.Equal(http.StatusOK, w.Code)
	var v models.Variant
	json.Unmarshal(w.Body.Bytes(), &v)
	return v
}
func (h *TestHelper) DeleteVariant(id uint) {
	w := h.PerformRequest("DELETE", fmt.Sprintf("/api/v1/stock/variants/%d", id), nil)
	h.Assert.Equal(http.StatusNoContent, w.Code)
}

// --- Characteristics ---
func (h *TestHelper) CreateCharacteristicType(payload gin.H) models.CharacteristicType {
	w := h.PerformRequest("POST", "/api/v1/stock/characteristics/types", payload)
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.CharacteristicType
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}
func (h *TestHelper) CreateCharacteristicValue(payload gin.H) models.CharacteristicValue {
	w := h.PerformRequest("POST", "/api/v1/stock/characteristics/values", payload)
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.CharacteristicValue
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// --- PriceTypes & Prices ---
func (h *TestHelper) CreatePriceType(name string) models.PriceType {
	w := h.PerformRequest("POST", "/api/v1/stock/price-types", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.PriceType
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}
func (h *TestHelper) GetPrice(variantID, priceTypeID uint) models.ItemPrice {
	path := fmt.Sprintf("/api/v1/stock/prices?item_id=%d&price_type_id=%d", variantID, priceTypeID)
	w := h.PerformRequest("GET", path, nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var price models.ItemPrice
	json.Unmarshal(w.Body.Bytes(), &price)
	return price
}

// --- Documents ---
func (h *TestHelper) CreateDocument(payload interface{}) models.Document {
	w := h.PerformRequest("POST", "/api/v1/stock/documents", payload)
	h.Assert.Equal(http.StatusCreated, w.Code, fmt.Sprintf("Failed to create document. Body: %s", w.Body.String()))
	var created models.Document
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}
func (h *TestHelper) GetDocument(docID uint) models.Document {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/documents/%d", docID), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var doc models.Document
	json.Unmarshal(w.Body.Bytes(), &doc)
	return doc
}
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
func (h *TestHelper) UpdateDocument(docID uint, payload interface{}) models.Document {
	w := h.PerformRequest("PUT", fmt.Sprintf("/api/v1/stock/documents/%d", docID), payload)
	h.Assert.Equal(http.StatusOK, w.Code)
	var doc models.Document
	json.Unmarshal(w.Body.Bytes(), &doc)
	return doc
}
func (h *TestHelper) DeleteDocument(docID uint) {
	w := h.PerformRequest("DELETE", fmt.Sprintf("/api/v1/stock/documents/%d", docID), nil)
	h.Assert.Equal(http.StatusNoContent, w.Code)
}
func (h *TestHelper) PostDocument(docID uint) {
	w := h.PerformRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/post", docID), nil)
	h.Assert.Equal(http.StatusOK, w.Code, fmt.Sprintf("Failed to post document. Body: %s", w.Body.String()))
}
func (h *TestHelper) CancelDocument(docID uint) {
	w := h.PerformRequest("POST", fmt.Sprintf("/api/v1/stock/documents/%d/cancel", docID), nil)
	h.Assert.Equal(http.StatusOK, w.Code, fmt.Sprintf("Failed to cancel document. Body: %s", w.Body.String()))
}

// --- Warehouses ---
func (h *TestHelper) CreateWarehouse(name string) models.Warehouse {
	w := h.PerformRequest("POST", "/api/v1/stock/warehouses", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}
func (h *TestHelper) GetWarehouse(id uint) models.Warehouse {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/warehouses/%d", id), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var wh models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &wh)
	return wh
}
func (h *TestHelper) ListWarehouses() []models.Warehouse {
	w := h.PerformRequest("GET", "/api/v1/stock/warehouses", nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var list []models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &list)
	return list
}
func (h *TestHelper) UpdateWarehouse(id uint, name string) models.Warehouse {
	w := h.PerformRequest("PUT", fmt.Sprintf("/api/v1/stock/warehouses/%d", id), gin.H{"name": name})
	h.Assert.Equal(http.StatusOK, w.Code)
	var updated models.Warehouse
	json.Unmarshal(w.Body.Bytes(), &updated)
	return updated
}
func (h *TestHelper) DeleteWarehouse(id uint) {
	w := h.PerformRequest("DELETE", fmt.Sprintf("/api/v1/stock/warehouses/%d", id), nil)
	h.Assert.Equal(http.StatusNoContent, w.Code)
}

// --- Categories ---
func (h *TestHelper) CreateCategory(name string) models.Category {
	w := h.PerformRequest("POST", "/api/v1/stock/categories", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.Category
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// --- Units ---
func (h *TestHelper) CreateUnit(name string) models.Unit {
	w := h.PerformRequest("POST", "/api/v1/stock/units", gin.H{"name": name})
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.Unit
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// --- Counterparties ---
func (h *TestHelper) CreateCounterparty(payload gin.H) models.Counterparty {
	w := h.PerformRequest("POST", "/api/v1/stock/counterparties", payload)
	h.Assert.Equal(http.StatusCreated, w.Code)
	var created models.Counterparty
	json.Unmarshal(w.Body.Bytes(), &created)
	return created
}

// --- Balances & Movements ---
func (h *TestHelper) GetBalances(warehouseID uint) []models.StockBalance {
	w := h.PerformRequest("GET", fmt.Sprintf("/api/v1/stock/balances/warehouse/%d", warehouseID), nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var balances []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balances)
	return balances
}
func (h *TestHelper) GetBalancesFiltered(warehouseID uint, queryParams string) []models.StockBalance {
	path := fmt.Sprintf("/api/v1/stock/balances/warehouse/%d", warehouseID)
	if queryParams != "" {
		path = fmt.Sprintf("%s?%s", path, queryParams)
	}
	w := h.PerformRequest("GET", path, nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var balances []models.StockBalance
	json.Unmarshal(w.Body.Bytes(), &balances)
	return balances
}
func (h *TestHelper) ListMovements() []models.StockMovement {
	w := h.PerformRequest("GET", "/api/v1/stock/movements", nil)
	h.Assert.Equal(http.StatusOK, w.Code)
	var movements []models.StockMovement
	json.Unmarshal(w.Body.Bytes(), &movements)
	return movements
}

func variantIDs(variants []models.Variant) []uint {
	ids := make([]uint, len(variants))
	for i, v := range variants {
		ids[i] = v.ID
	}
	return ids
}
