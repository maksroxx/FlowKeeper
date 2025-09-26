package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type WarehouseHandler struct {
	svc service.WarehouseService
}

func NewWarehouseHandler(svc service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{svc: svc}
}

func (h *WarehouseHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/warehouses")
	{
		grp.POST("/", h.Create)
		grp.GET("/", h.List)
		grp.GET("/:id", h.GetByID)
	}
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var w stock.Warehouse
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.Create(w.Name, w.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *WarehouseHandler) List(c *gin.Context) {
	ws, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ws)
}

func (h *WarehouseHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	w, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}
