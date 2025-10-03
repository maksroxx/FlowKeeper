package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type MovementHandler struct {
	service service.StockMovementService
}

func NewMovementHandler(s service.StockMovementService) *MovementHandler {
	return &MovementHandler{service: s}
}

func (h *MovementHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/movements")
	{
		grp.GET("", h.List)
		grp.GET("/:id", h.GetByID)
	}
}

func (h *MovementHandler) List(c *gin.Context) {
	movementDTOs, err := h.service.ListAsDTO()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, movementDTOs)
}

func (h *MovementHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	movement, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movement not found"})
		return
	}
	c.JSON(http.StatusOK, movement)
}
