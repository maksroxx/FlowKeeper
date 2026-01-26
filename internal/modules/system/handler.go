package system

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/db"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	dbConf config.DatabaseConfig
}

func NewHandler(d *gorm.DB, conf config.DatabaseConfig) *Handler {
	return &Handler{db: d, dbConf: conf}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/backup", h.CreateBackup)
	r.POST("/restore", h.RestoreBackup)
}

func (h *Handler) CreateBackup(c *gin.Context) {
	tempName := fmt.Sprintf("backup_%d.db", time.Now().Unix())
	tempPath := filepath.Join(os.TempDir(), tempName)

	if err := h.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", tempPath)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backup: " + err.Error()})
		return
	}
	defer os.Remove(tempPath)

	c.FileAttachment(tempPath, fmt.Sprintf("flowkeeper_backup_%s.db", time.Now().Format("2006-01-02_15-04")))
}

func (h *Handler) RestoreBackup(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	currentDBPath := h.dbConf.DSN
	if !filepath.IsAbs(currentDBPath) {
		cwd, _ := os.Getwd()
		currentDBPath = filepath.Join(cwd, currentDBPath)
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get DB instance"})
		return
	}
	if err := sqlDB.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close DB connection: " + err.Error()})
		return
	}

	tempRestorePath := currentDBPath + ".restore_tmp"
	if err := c.SaveUploadedFile(file, tempRestorePath); err != nil {
		reconnect(h)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save upload: " + err.Error()})
		return
	}

	os.Rename(currentDBPath, currentDBPath+".bak")

	if err := os.Rename(tempRestorePath, currentDBPath); err != nil {
		os.Rename(currentDBPath+".bak", currentDBPath)
		reconnect(h)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to replace DB file: " + err.Error()})
		return
	}

	if err := reconnect(h); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "restored_restart_required", "message": "База восстановлена. Пожалуйста, перезапустите приложение."})
		return
	}
	os.Remove(currentDBPath + ".bak")
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "База успешно восстановлена"})
}

func reconnect(h *Handler) error {
	newDB, err := db.Connect(h.dbConf)
	if err != nil {
		return err
	}
	*h.db = *newDB
	return nil
}
