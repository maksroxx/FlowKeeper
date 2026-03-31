package shop

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(r Repository) *Handler {
	return &Handler{repo: r}
}

func (h *Handler) GetCatalog(c *gin.Context) {
	s, err := h.repo.GetSettings()
	if err != nil || !s.IsEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Shop is disabled"})
		return
	}

	search := c.Query("search")
	items, err := h.repo.GetCatalog(search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetProduct(c *gin.Context) {
	s, err := h.repo.GetSettings()
	if err != nil || !s.IsEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Shop is disabled"})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	product, err := h.repo.GetProductDetails(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *Handler) GetSettings(c *gin.Context) {
	s, err := h.repo.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var s ShopSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateSettings(&s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) ToggleStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		IsPublic   *bool `json:"is_public"`
		IsFeatured *bool `json:"is_featured"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.ToggleProductStatus(uint(id), req.IsPublic, req.IsFeatured); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
