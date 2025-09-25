package users

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string { return "users" }

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)

	group := r.Group("/users")
	{
		group.POST("", func(c *gin.Context) {
			var req struct {
				Name     string `json:"name"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Role     string `json:"role"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user, err := service.AddUser(req.Name, req.Email, req.Password, req.Role)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, user)
		})

		group.GET("", func(c *gin.Context) {
			users, err := service.ListUsers()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, users)
		})

		group.GET("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			user, err := service.GetUser(uint(id))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusOK, user)
		})

		group.PUT("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			var req struct {
				Name     string `json:"name"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Role     string `json:"role"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user := &User{
				ID:       uint(id),
				Name:     req.Name,
				Email:    req.Email,
				Password: req.Password,
				Role:     req.Role,
			}
			if err := service.UpdateUser(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, user)
		})

		group.DELETE("/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			if err := service.DeleteUser(uint(id)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"deleted": true})
		})
	}
}
