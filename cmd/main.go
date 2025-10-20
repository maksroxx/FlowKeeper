package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/api"
	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/core"
	"github.com/maksroxx/flowkeeper/internal/db"
	"github.com/maksroxx/flowkeeper/internal/modules/stock"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.RedirectTrailingSlash = true
	config := cors.DefaultConfig()
	// Разрешаем все методы
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	// Разрешаем все заголовки
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	// Разрешаем передачу cookie и токенов
	config.AllowCredentials = true
	// Устанавливаем время кеширования
	config.MaxAge = 12 * time.Hour

	// САМАЯ ВАЖНАЯ ЧАСТЬ: Разрешаем любой origin.
	// Вместо AllowOrigins = []string{"*"}, что несовместимо с Credentials,
	// мы используем функцию, которая проверяет и разрешает любой источник.
	config.AllowOriginFunc = func(origin string) bool {
		_, err := url.Parse(origin)
		return err == nil // Если origin - это валидный URL, разрешаем его.
	}

	r.Use(cors.New(config))
	app := core.NewApp(database, r)

	if cfg.Modules.Stock {
		app.RegisterModule(stock.NewModule())
	}

	// if cfg.Modules.Users {
	// 	app.RegisterModule(users.NewModule())
	// }

	if err := app.Migrate(); err != nil {
		log.Fatal(err)
	}

	api.InitAPI(r, app)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
