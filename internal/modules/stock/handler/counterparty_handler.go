package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type CounterpartyHandler struct {
	service service.CounterpartyService
}

func NewCounterpartyHandler(s service.CounterpartyService) *CounterpartyHandler {
	return &CounterpartyHandler{service: s}
}

func (h *CounterpartyHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/counterparties")
	{
		grp.POST("", h.Create)
		grp.GET("", h.Search)
		grp.GET("/:id", h.GetByID)
		grp.PUT("/:id", h.Update)
		grp.DELETE("/:id", h.Delete)
	}
}

func (h *CounterpartyHandler) Create(c *gin.Context) {
	var cp models.Counterparty
	if err := c.ShouldBindJSON(&cp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdCounterparty, err := h.service.Create(&cp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdCounterparty)
}

func (h *CounterpartyHandler) List(c *gin.Context) {
	counterparties, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, counterparties)
}

func (h *CounterpartyHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	counterparty, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Counterparty not found"})
		return
	}
	c.JSON(http.StatusOK, counterparty)
}

func (h *CounterpartyHandler) Update(c *gin.Context) {
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

	updatedCounterparty, err := h.service.Update(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedCounterparty)
}

func (h *CounterpartyHandler) Delete(c *gin.Context) {
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

func (h *CounterpartyHandler) Search(c *gin.Context) {
	var filter models.CounterpartyFilter

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		filter.Offset = offset
	}

	counterparties, err := h.service.Search(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, counterparties)
}
