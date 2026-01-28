package bootstrap

import (
	"errors"
	"log"

	"github.com/maksroxx/flowkeeper/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	log.Println("--- [BOOTSTRAP] Начало проверки данных ---")

	if err := db.AutoMigrate(&users.Role{}, &users.User{}); err != nil {
		log.Printf("❌ [BOOTSTRAP] Ошибка миграции таблиц пользователей: %v", err)
		return
	}

	var adminRole users.Role
	err := db.Where("name = ?", "admin").First(&adminRole).Error

	if err == nil {
		var adminCount int64
		db.Model(&users.User{}).Where("role_id = ?", adminRole.ID).Count(&adminCount)

		if adminCount > 0 {
			log.Println("✅ [BOOTSTRAP] Администратор найден. Сидинг пропущен.")
			return
		}
		log.Println("⚠️ [BOOTSTRAP] Роли есть, но админов нет. Создаем админа...")
	} else {
		log.Println("⚡ [BOOTSTRAP] Чистая база. Запуск полного сидинга...")
	}

	log.Println("⚡ [BOOTSTRAP] База пуста. Создаем структуру ролей и пользователей...")

	adminRole = users.Role{Name: "admin", Permissions: []string{"all"}}
	if err := ensureRole(db, &adminRole); err != nil {
		log.Printf("❌ [BOOTSTRAP] Не удалось создать роль admin: %v", err)
		return
	}

	workerRole := users.Role{Name: "worker", Permissions: []string{"view_dashboard", "view_stock", "create_document"}}
	if err := ensureRole(db, &workerRole); err != nil {
		log.Printf("❌ [BOOTSTRAP] Не удалось создать роль worker: %v", err)
	}

	managerRole := users.Role{Name: "manager", Permissions: []string{"view_dashboard", "view_stock", "view_reports", "create_document", "approve_document"}}
	if err := ensureRole(db, &managerRole); err != nil {
		log.Printf("❌ [BOOTSTRAP] Не удалось создать роль manager: %v", err)
	}

	err = createUser(db, "Super Admin", "admin@flowkeeper.com", "admin", adminRole.ID)
	if err != nil {
		log.Printf("❌ [BOOTSTRAP] КРИТИЧЕСКАЯ ОШИБКА: Не удалось создать админа: %v", err)
	} else {
		log.Println("✅ [BOOTSTRAP] Супер Админ успешно создан (admin@flowkeeper.com / admin)")
	}

	createUser(db, "Sklad", "worker@flowkeeper.com", "user", workerRole.ID)

	log.Println("--- [BOOTSTRAP] Завершено ---")
}

func ensureRole(db *gorm.DB, role *users.Role) error {
	var existing users.Role
	err := db.Where("name = ?", role.Name).First(&existing).Error

	if err == nil {
		role.ID = existing.ID
		log.Printf("   -> Роль '%s' найдена (ID: %d)", role.Name, role.ID)
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.Create(role).Error; createErr != nil {
			return createErr
		}
		log.Printf("   -> Роль '%s' создана (ID: %d)", role.Name, role.ID)
		return nil
	}

	return err
}

func createUser(db *gorm.DB, name, email, password string, roleID uint) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := users.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		RoleID:   roleID,
	}

	if err := db.Create(&user).Error; err != nil {
		return err
	}
	log.Printf("   -> Пользователь '%s' создан", email)
	return nil
}
