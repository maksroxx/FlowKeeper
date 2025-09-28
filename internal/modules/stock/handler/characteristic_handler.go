package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type CharacteristicHandler struct {
	service service.CharacteristicService
}

func NewCharacteristicHandler(s service.CharacteristicService) *CharacteristicHandler {
	return &CharacteristicHandler{service: s}
}

func (h *CharacteristicHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/characteristics")
	{
		// Роуты для типов характеристик
		types := grp.Group("/types")
		types.POST("", h.CreateType)
		types.GET("", h.ListTypes)
		types.GET("/:id", h.GetTypeByID)
		types.PUT("/:id", h.UpdateType)
		types.DELETE("/:id", h.DeleteType)

		// Роуты для характеристик
		values := grp.Group("/values")
		values.POST("", h.CreateValue)
		values.GET("", h.ListValues)
		values.GET("/:id", h.GetValueByID)
		values.PUT("/:id", h.UpdateValue)
		values.DELETE("/:id", h.DeleteValue)
	}
}

func (h *CharacteristicHandler) CreateType(c *gin.Context) {
	var ct models.CharacteristicType
	if err := c.ShouldBindJSON(&ct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.service.CreateType(&ct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *CharacteristicHandler) ListTypes(c *gin.Context) {
	list, err := h.service.ListTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *CharacteristicHandler) GetTypeByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	ct, err := h.service.GetTypeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Characteristic type not found"})
		return
	}
	c.JSON(http.StatusOK, ct)
}

func (h *CharacteristicHandler) UpdateType(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ct models.CharacteristicType
	if err := c.ShouldBindJSON(&ct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ct.ID = uint(id)
	updated, err := h.service.UpdateType(&ct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *CharacteristicHandler) DeleteType(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.DeleteType(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CharacteristicHandler) CreateValue(c *gin.Context) {
	var cv models.CharacteristicValue
	if err := c.ShouldBindJSON(&cv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.service.CreateValue(&cv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *CharacteristicHandler) ListValues(c *gin.Context) {
	list, err := h.service.ListValues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *CharacteristicHandler) GetValueByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	cv, err := h.service.GetValueByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Characteristic value not found"})
		return
	}
	c.JSON(http.StatusOK, cv)
}

func (h *CharacteristicHandler) UpdateValue(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var cv models.CharacteristicValue
	if err := c.ShouldBindJSON(&cv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cv.ID = uint(id)
	updated, err := h.service.UpdateValue(&cv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *CharacteristicHandler) DeleteValue(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.DeleteValue(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
