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
	"github.com/maksroxx/flowkeeper/internal/modules/analytics"
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
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	config.AllowOriginFunc = func(origin string) bool {
		_, err := url.Parse(origin)
		return err == nil
	}

	r.Use(cors.New(config))
	app := core.NewApp(database, r)

	if cfg.Modules.Stock {
		app.RegisterModule(stock.NewModule())
	}

	if cfg.Modules.Analytics {
		app.RegisterModule(analytics.NewModule())
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
