package main

import (
	"fmt"
	"log"

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
