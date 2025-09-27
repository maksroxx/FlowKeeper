package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type PriceHandler struct {
	service service.PriceService
}

func NewPriceHandler(s service.PriceService) *PriceHandler {
	return &PriceHandler{service: s}
}

func (h *PriceHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/prices")
	{
		grp.GET("", h.GetItemPrice)
	}
}

func (h *PriceHandler) GetItemPrice(c *gin.Context) {
	itemIDStr := c.Query("item_id")
	priceTypeIDStr := c.Query("price_type_id")

	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item_id parameter"})
		return
	}

	priceTypeID, err := strconv.ParseUint(priceTypeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price_type_id parameter"})
		return
	}

	price, err := h.service.GetPrice(uint(itemID), uint(priceTypeID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if price == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Price not found for this item and price type"})
		return
	}

	c.JSON(http.StatusOK, price)
}
