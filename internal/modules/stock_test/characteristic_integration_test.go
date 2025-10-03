package stocktest

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCharacteristic_Creation_Integration(t *testing.T) {
	router, db := setupTestRouter("characteristic_db")
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	h := NewTestHelper(t, router)

	t.Log("--- Тестирование создания CharacteristicType ---")

	charTypeColor := h.CreateCharacteristicType(gin.H{"name": "Цвет"})
	h.Assert.NotZero(charTypeColor.ID, "ID для 'Цвет' не должен быть нулевым")
	h.Assert.Equal("Цвет", charTypeColor.Name)

	charTypeSize := h.CreateCharacteristicType(gin.H{"name": "Размер"})
	h.Assert.NotZero(charTypeSize.ID, "ID для 'Размер' не должен быть нулевым")
	h.Assert.Equal("Размер", charTypeSize.Name)

	t.Log("--- Тестирование создания CharacteristicValue ---")

	charValueRed := h.CreateCharacteristicValue(gin.H{
		"characteristic_type_id": charTypeColor.ID,
		"value":                  "Красный",
	})
	h.Assert.NotZero(charValueRed.ID)
	h.Assert.Equal(charTypeColor.ID, charValueRed.CharacteristicTypeID)
	h.Assert.Equal("Красный", charValueRed.Value)

	charValueBlue := h.CreateCharacteristicValue(gin.H{
		"characteristic_type_id": charTypeColor.ID,
		"value":                  "Синий",
	})
	h.Assert.Equal(charTypeColor.ID, charValueBlue.CharacteristicTypeID)
	h.Assert.Equal("Синий", charValueBlue.Value)

	charValueL := h.CreateCharacteristicValue(gin.H{
		"characteristic_type_id": charTypeSize.ID,
		"value":                  "L",
	})
	h.Assert.Equal(charTypeSize.ID, charValueL.CharacteristicTypeID)
	h.Assert.Equal("L", charValueL.Value)

	t.Log("API для создания характеристик работает корректно.")
}
