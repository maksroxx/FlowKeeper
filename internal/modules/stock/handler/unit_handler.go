package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type UnitHandler struct {
	svc service.UnitService
}

func NewUnitHandler(svc service.UnitService) *UnitHandler {
	return &UnitHandler{svc: svc}
}

func (h *UnitHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/units")
	{
		grp.POST("/", h.Create)
		grp.GET("/", h.List)
		grp.GET("/:id", h.GetByID)
	}
}

func (h *UnitHandler) Create(c *gin.Context) {
	var u stock.Unit
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.Create(u.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *UnitHandler) List(c *gin.Context) {
	units, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, units)
}

func (h *UnitHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	unit, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, unit)
}
