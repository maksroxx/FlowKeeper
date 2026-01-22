package main

import (
	"flag"
	"log"

	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/db"
	"github.com/maksroxx/flowkeeper/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	log.Println("Запуск сидера авторизации...")

	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфига: %v", err)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}

	log.Println("Проверка таблиц...")
	if err := database.AutoMigrate(&users.Role{}, &users.User{}); err != nil {
		log.Fatalf("❌ Ошибка миграции: %v", err)
	}

	seedRoles(database)
	seedUsers(database)

	log.Println("✅ Сидинг авторизации успешно завершен!")
}

func seedRoles(db *gorm.DB) {
	roles := []users.Role{
		{
			Name:        "admin",
			Permissions: []string{"all"},
		},
		{
			Name: "worker",
			Permissions: []string{
				"view_dashboard",
				"view_stock",
				"create_document",
				"view_inventory",
			},
		},
		{
			Name: "manager",
			Permissions: []string{
				"view_dashboard",
				"view_stock",
				"view_reports",
				"view_counterparties",
				"create_document",
				"approve_document",
			},
		},
	}

	for _, r := range roles {
		var existing users.Role
		if err := db.Where("name = ?", r.Name).First(&existing).Error; err != nil {
			if err := db.Create(&r).Error; err != nil {
				log.Printf("Ошибка создания роли '%s': %v", r.Name, err)
			} else {
				log.Printf("🔹 Создана роль: %s", r.Name)
			}
		} else {
			existing.Permissions = r.Permissions
			db.Save(&existing)
			log.Printf("🔸 Роль '%s' уже существует (права обновлены)", r.Name)
		}
	}
}

func seedUsers(db *gorm.DB) {
	var adminRole, workerRole users.Role
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "worker").First(&workerRole)

	usersList := []struct {
		Name     string
		Email    string
		Password string
		RoleID   uint
	}{
		{
			Name:     "Super Admin",
			Email:    "admin@flowkeeper.com",
			Password: "admin",
			RoleID:   adminRole.ID,
		},
		{
			Name:     "Ivan Sklad",
			Email:    "worker@flowkeeper.com",
			Password: "user",
			RoleID:   workerRole.ID,
		},
	}

	for _, u := range usersList {
		var existing users.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err != nil {
			hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

			user := users.User{
				Name:     u.Name,
				Email:    u.Email,
				Password: string(hash),
				RoleID:   u.RoleID,
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("⚠️ Ошибка создания пользователя '%s': %v", u.Email, err)
			} else {
				log.Printf("👤 Создан пользователь: %s (Пароль: %s)", u.Email, u.Password)
			}
		} else {
			log.Printf("OK Пользователь '%s' уже существует", u.Email)
		}
	}
}
