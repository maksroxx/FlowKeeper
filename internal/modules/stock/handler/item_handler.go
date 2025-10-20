package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type VariantHandler struct {
	service          service.VariantService
	inventoryService service.InventoryService
}

func NewVariantHandler(s service.VariantService, inv service.InventoryService) *VariantHandler {
	return &VariantHandler{service: s, inventoryService: inv}
}

func (h *VariantHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/variants")
	{
		grp.POST("", h.Create)
		// grp.GET("", h.List)
		grp.GET("/:id", h.GetByID)
		grp.PUT("/:id", h.Update)
		grp.DELETE("/:id", h.Delete)
		grp.GET("", h.Search)
		grp.GET("/:id/stock", h.GetStock)
	}
}

func (h *VariantHandler) Create(c *gin.Context) {
	var v models.Variant
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.service.Create(&v)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *VariantHandler) List(c *gin.Context) {
	list, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *VariantHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	variantDTO, err := h.service.GetByIDAsDTO(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch variant data"})
		return
	}
	if variantDTO == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	c.JSON(http.StatusOK, variantDTO)
}

func (h *VariantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedVariant, err := h.service.Update(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedVariant)
}

func (h *VariantHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *VariantHandler) Search(c *gin.Context) {
	var filter models.VariantFilter

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if catID, err := strconv.ParseUint(catIDStr, 10, 64); err == nil {
			catIDUint := uint(catID)
			filter.CategoryID = &catIDUint
		}
	}

	if sku := c.Query("sku"); sku != "" {
		filter.SKU = &sku
	}

	filter.StockStatus = c.DefaultQuery("stock_status", "all")
	if whIDStr := c.Query("warehouse_id"); whIDStr != "" {
		if whID, err := strconv.ParseUint(whIDStr, 10, 64); err == nil {
			whIDUint := uint(whID)
			filter.WarehouseID = &whIDUint
		}
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		filter.Offset = offset
	}

	variants, err := h.service.Search(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variants)
}

func (h *VariantHandler) GetStock(c *gin.Context) {
	variantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Variant ID"})
		return
	}

	stockLevels, err := h.inventoryService.GetStockByVariant(uint(variantID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stockLevels)
}
