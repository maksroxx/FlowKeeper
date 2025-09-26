package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type CounterpartyHandler struct {
	svc service.CounterpartyService
}

func NewCounterpartyHandler(svc service.CounterpartyService) *CounterpartyHandler {
	return &CounterpartyHandler{svc: svc}
}

func (h *CounterpartyHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/counterparties")
	{
		grp.POST("/", h.Create)
		grp.GET("/", h.List)
		grp.GET("/:id", h.GetByID)
	}
}

func (h *CounterpartyHandler) Create(c *gin.Context) {
	var cp stock.Counterparty
	if err := c.ShouldBindJSON(&cp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.Create(cp.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *CounterpartyHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *CounterpartyHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	cp, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cp)
}
