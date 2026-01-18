package files

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания папки uploads: %v\n", err)
	}
	return &Handler{}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("/upload", h.UploadFile)
}

func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))

	dst := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file: %v", err)})
		return
	}

	publicURL := "/uploads/" + filename
	c.JSON(http.StatusOK, gin.H{"url": publicURL})
}
