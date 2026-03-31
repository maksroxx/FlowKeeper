package system

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/db"
	"gorm.io/driver/postgres"
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
	r.POST("/config/db", h.SwitchDatabase)
}

func (h *Handler) CreateBackup(c *gin.Context) {
	if h.dbConf.Driver != "sqlite" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Бэкап доступен только для SQLite"})
		return
	}

	tempName := fmt.Sprintf("backup_%d.db", time.Now().Unix())
	tempPath := filepath.Join(os.TempDir(), tempName)

	if err := h.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", tempPath)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания бэкапа: " + err.Error()})
		return
	}
	defer os.Remove(tempPath)

	fileName := fmt.Sprintf("FlowKeeper_Backup_%s.db", time.Now().Format("2006-01-02_15-04"))
	c.FileAttachment(tempPath, fileName)
}

func (h *Handler) RestoreBackup(c *gin.Context) {
	if h.dbConf.Driver != "sqlite" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Восстановление доступно только для SQLite"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}

	dbPath := h.dbConf.DSN
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить доступ к БД"})
		return
	}
	if err := sqlDB.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось закрыть соединение: " + err.Error()})
		return
	}

	backupPath := dbPath + ".bak"
	os.Remove(backupPath)
	if err := os.Rename(dbPath, backupPath); err != nil {
		if !os.IsNotExist(err) {
			_ = reconnect(h)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка бэкапа текущей версии: " + err.Error()})
			return
		}
	}

	src, err := file.Open()
	if err != nil {
		os.Rename(backupPath, dbPath)
		reconnect(h)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения файла: " + err.Error()})
		return
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		os.Rename(backupPath, dbPath)
		reconnect(h)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания файла БД: " + err.Error()})
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		os.Rename(backupPath, dbPath)
		reconnect(h)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка записи файла: " + err.Error()})
		return
	}

	if err := reconnect(h); err != nil {
		os.Remove(dbPath)
		os.Rename(backupPath, dbPath)
		reconnect(h)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Загруженный файл не является валидной БД или поврежден. Восстановлена прежняя версия."})
		return
	}

	os.Remove(backupPath)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "База успешно восстановлена. Пожалуйста, перезайдите в систему.",
	})
}

func reconnect(h *Handler) error {
	newDB, err := db.Connect(h.dbConf)
	if err != nil {
		return err
	}
	*h.db = *newDB
	return nil
}

type DBConnectionRequest struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

func (h *Handler) SwitchDatabase(c *gin.Context) {
	var req DBConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var newDSN string

	if req.Driver == "postgres" {
		if req.SSLMode == "" {
			req.SSLMode = "disable"
		}
		newDSN = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			req.Host, req.User, req.Password, req.DBName, req.Port, req.SSLMode)

		testDB, err := gorm.Open(postgres.Open(newDSN), &gorm.Config{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось подключиться к PostgreSQL: " + err.Error()})
			return
		}
		sqlDB, _ := testDB.DB()
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Соединение установлено, но пинг не прошел: " + err.Error()})
			return
		}
		sqlDB.Close()

	} else {
		req.Driver = "sqlite"
		newDSN = "./data/local.db"
	}

	configPath := "config/config.yaml"

	contentBytes, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать config.yaml"})
		return
	}
	content := string(contentBytes)

	content = regexp.MustCompile(`driver: ".*"`).ReplaceAllString(content, fmt.Sprintf(`driver: "%s"`, req.Driver))

	var cfgMap map[string]interface{}
	if err := yaml.Unmarshal(contentBytes, &cfgMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга конфига"})
		return
	}

	dbMap := cfgMap["database"].(map[string]interface{})
	dbMap["driver"] = req.Driver
	dbMap["dsn"] = newDSN

	newBytes, err := yaml.Marshal(cfgMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка формирования конфига"})
		return
	}

	if err := os.WriteFile(configPath, newBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось записать config.yaml"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Настройки сохранены. Пожалуйста, перезапустите приложение."})
}
