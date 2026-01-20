package files

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

func CleanupOrphanedImages(db *gorm.DB) {
	uploadDir := "uploads"

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return
	}

	var dbUrls []string
	if err := db.Table("product_images").Pluck("url", &dbUrls).Error; err != nil {
		log.Printf("Ошибка очистки файлов: не удалось прочитать БД: %v", err)
		return
	}

	validFiles := make(map[string]bool)
	for _, url := range dbUrls {
		filename := filepath.Base(url)
		validFiles[filename] = true
	}

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Printf("Ошибка чтения папки uploads: %v", err)
		return
	}

	deletedCount := 0
	var sizeFreed int64 = 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasPrefix(name, ".") {
			continue
		}

		if !validFiles[name] {
			fullPath := filepath.Join(uploadDir, name)

			info, _ := entry.Info()
			if info != nil {
				sizeFreed += info.Size()
			}

			if err := os.Remove(fullPath); err != nil {
				log.Printf("Не удалось удалить мусорный файл %s: %v", name, err)
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		mbFreed := float64(sizeFreed) / 1024 / 1024
		log.Printf("Очистка мусора: удалено %d файлов (освобождено %.2f MB)", deletedCount, mbFreed)
	} else {
		log.Println("Очистка мусора: лишних файлов не найдено.")
	}
}
