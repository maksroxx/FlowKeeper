package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/api"
	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/core"
	"github.com/maksroxx/flowkeeper/internal/db"
	"github.com/maksroxx/flowkeeper/internal/modules/analytics"
	"github.com/maksroxx/flowkeeper/internal/modules/files"
	"github.com/maksroxx/flowkeeper/internal/modules/reports"
	"github.com/maksroxx/flowkeeper/internal/modules/stock"
	"github.com/maksroxx/flowkeeper/internal/modules/users"
)

func main() {
	fontDir := "assets/fonts"
	if err := os.MkdirAll(fontDir, 0755); err != nil {
		panic(err)
	}

	fonts := map[string]string{
		"Roboto-Regular.ttf": "https://github.com/googlefonts/roboto/raw/main/src/hinted/Roboto-Regular.ttf",
		"Roboto-Bold.ttf":    "https://github.com/googlefonts/roboto/raw/main/src/hinted/Roboto-Bold.ttf",
	}

	for name, url := range fonts {
		path := filepath.Join(fontDir, name)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Remove(path + ".json")
			os.Remove(path + ".z")

			fmt.Printf("Скачивание %s...\n", name)

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("⚠️ Не удалось скачать шрифт %s: %v", name, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				log.Printf("⚠️ Ошибка сервера при скачивании шрифта %s: %d", name, resp.StatusCode)
				continue
			}

			out, err := os.Create(path)
			if err != nil {
				log.Printf("⚠️ Ошибка создания файла %s: %v", name, err)
				continue
			}
			defer out.Close()

			size, err := io.Copy(out, resp.Body)
			if err != nil {
				log.Printf("⚠️ Ошибка сохранения файла %s: %v", name, err)
				continue
			}

			fmt.Printf("✅ Успешно скачан: %s (%.2f KB)\n", name, float64(size)/1024)
		} else {
		}
	}

	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.RedirectTrailingSlash = true
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	corsConfig.AllowOriginFunc = func(origin string) bool {
		_, err := url.Parse(origin)
		return err == nil
	}

	r.Use(cors.New(corsConfig))
	app := core.NewApp(database, r)

	if cfg.Modules.Files {
		app.RegisterModule(files.NewModule())
	}

	if cfg.Modules.Stock {
		app.RegisterModule(stock.NewModule(cfg.Auth))
	}

	if cfg.Modules.Users {
		app.RegisterModule(users.NewModule(cfg.Auth))
	}

	if cfg.Modules.Analytics {
		app.RegisterModule(analytics.NewModule())
	}

	if cfg.Modules.Reports {
		app.RegisterModule(reports.NewModule())
	}

	if err := app.Migrate(); err != nil {
		log.Fatal(err)
	}

	files.CleanupOrphanedImages(database)
	api.InitAPI(r, app)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Сервер запущен на %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
