package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	// /api/v1/analytics
	r.GET("/dashboard", h.GetDashboard)
}

func (h *Handler) GetDashboard(c *gin.Context) {
	var warehouseID *uint
	if idStr := c.Query("warehouse_id"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			uid := uint(id)
			warehouseID = &uid
		}
	}

	data, err := h.service.GetDashboardData(warehouseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
