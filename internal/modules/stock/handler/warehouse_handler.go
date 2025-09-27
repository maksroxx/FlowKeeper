package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type WarehouseHandler struct {
	service service.WarehouseService
}

func NewWarehouseHandler(s service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{service: s}
}

func (h *WarehouseHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/warehouses")
	{
		grp.POST("", h.Create)
		grp.GET("", h.List)
		grp.GET("/:id", h.GetByID)
		grp.PUT("/:id", h.Update)
		grp.DELETE("/:id", h.Delete)
	}
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var wh models.Warehouse
	if err := c.ShouldBindJSON(&wh); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdWarehouse, err := h.service.Create(wh.Name, wh.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdWarehouse)
}

func (h *WarehouseHandler) List(c *gin.Context) {
	warehouses, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, warehouses)
}

func (h *WarehouseHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	warehouse, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}
	c.JSON(http.StatusOK, warehouse)
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var wh models.Warehouse
	if err := c.ShouldBindJSON(&wh); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wh.ID = uint(id)

	updatedWarehouse, err := h.service.Update(&wh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedWarehouse)
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
