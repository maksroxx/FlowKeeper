package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type BalanceHandler struct {
	service service.InventoryService
}

func NewBalanceHandler(s service.InventoryService) *BalanceHandler {
	return &BalanceHandler{service: s}
}

func (h *BalanceHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/balances")
	{
		grp.GET("/warehouse/:id", h.GetByWarehouse)
	}
}

func (h *BalanceHandler) GetByWarehouse(c *gin.Context) {
	warehouseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Warehouse ID"})
		return
	}

	var filter models.StockFilter
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, err := strconv.ParseUint(catIDStr, 10, 32)
		if err == nil {
			catIDUint := uint(catID)
			filter.CategoryID = &catIDUint
		}
	}
	if sku := c.Query("sku"); sku != "" {
		filter.SKU = &sku
	}

	if minQtyStr := c.Query("min_qty"); minQtyStr != "" {
		minQty, err := decimal.NewFromString(minQtyStr)
		if err == nil {
			filter.MinQty = &minQty
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_qty format"})
			return
		}
	}

	balances, err := h.service.ListByWarehouseFiltered(uint(warehouseID), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balances)
}
