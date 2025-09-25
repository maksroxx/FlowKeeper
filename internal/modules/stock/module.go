package stock

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string { return "stock" }

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Item{}, &StockMovement{})
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := &itemRepository{db: db}
	service := &stockService{repo: repo}

	group := r.Group("/stock")
	{
		group.POST("/items", func(c *gin.Context) {
			var req struct {
				Name  string `json:"name"`
				Stock int    `json:"stock"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			item, err := service.AddItem(req.Name, req.Stock)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, item)
		})

		group.GET("/items", func(c *gin.Context) {
			items, err := service.ListItems()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, items)
		})
	}
}
