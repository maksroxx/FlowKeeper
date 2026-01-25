package bootstrap

import (
	"log"

	"github.com/maksroxx/flowkeeper/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	var count int64
	db.Model(&users.User{}).Count(&count)

	if count > 0 {
		log.Println("Сидинг не требуется: пользователи уже есть.")
		return
	}

	log.Println("База пуста. Запуск начального сидинга...")

	roles := []users.Role{
		{Name: "admin", Permissions: []string{"all"}},
		{Name: "worker", Permissions: []string{"view_dashboard", "view_stock", "create_document"}},
		{Name: "manager", Permissions: []string{"view_dashboard", "view_stock", "view_reports", "create_document", "approve_document"}},
	}

	for _, r := range roles {
		if err := db.FirstOrCreate(&r, users.Role{Name: r.Name}).Error; err != nil {
			log.Printf("Ошибка создания роли %s: %v", r.Name, err)
		}
	}

	var adminRole, workerRole users.Role
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "worker").First(&workerRole)

	usersList := []struct {
		Name     string
		Email    string
		Password string
		RoleID   uint
	}{
		{"Super Admin", "admin@sklad.com", "admin", adminRole.ID},
		{"Sklad", "user@sklad.com", "user", workerRole.ID},
	}

	for _, u := range usersList {
		hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		user := users.User{
			Name:     u.Name,
			Email:    u.Email,
			Password: string(hash),
			RoleID:   u.RoleID,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Ошибка создания %s: %v", u.Email, err)
		}
	}

	log.Println("✅ Сидинг завершен успешно.")
}
