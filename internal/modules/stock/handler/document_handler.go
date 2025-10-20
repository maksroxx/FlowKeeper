package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type DocumentHandler struct {
	service service.DocumentService
}

func NewDocumentHandler(s service.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: s}
}

func (h *DocumentHandler) Register(r *gin.RouterGroup) {
	grp := r.Group("/documents")
	{
		grp.POST("", h.Create)
		grp.GET("", h.List)
		grp.GET("/:id", h.GetByID)
		grp.PUT("/:id", h.Update)
		grp.DELETE("/:id", h.Delete)

		grp.POST("/:id/post", h.Post)     // провести документ
		grp.POST("/:id/cancel", h.Cancel) // отменить
	}
}

func (h *DocumentHandler) Create(c *gin.Context) {
	var doc models.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdDoc, err := h.service.Create(&doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdDoc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	var filter models.DocumentFilter

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if types := c.QueryArray("types"); len(types) > 0 {
		filter.Types = types
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if t, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filter.DateFrom = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, use RFC3339"})
			return
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if t, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filter.DateTo = &t
		}
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(c.Query("offset")); err == nil {
		filter.Offset = offset
	}

	docDTOs, err := h.service.SearchAsDTO(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, docDTOs)
}

func (h *DocumentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	doc, err := h.service.GetByIDAsDTO(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var updatePayload struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&updatePayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	docToUpdate := &models.Document{
		Comment: updatePayload.Comment,
	}

	updatedDoc, err := h.service.Update(uint(id), docToUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedDoc)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
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

func (h *DocumentHandler) Post(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	if err := h.service.Post(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Document posted successfully"})
}

func (h *DocumentHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	if err := h.service.Cancel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Document canceled successfully"})
}
