package service

import (
	"strings"
	"testing"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockTxManager struct{}

func (m *mockTxManager) DoInTx(fn func(tx *gorm.DB) error) error {
	return fn(nil)
}

type mockSequenceRepo struct {
	currentNumber uint
}

func (m *mockSequenceRepo) GetNext(tx *gorm.DB, id string) (*models.DocumentSequence, error) {
	m.currentNumber++
	return &models.DocumentSequence{ID: id, LastNumber: m.currentNumber}, nil
}

type mockDocumentRepo struct {
	repository.DocumentRepository
	existingNumbers map[string]bool
}

func (m *mockDocumentRepo) GetByNumber(number string) (*models.Document, error) {
	if m.existingNumbers[number] {
		return &models.Document{Number: number}, nil
	}
	return nil, nil
}

func TestSequenceService_GenerateNextDocumentNumber(t *testing.T) {
	txManager := &mockTxManager{}
	seqRepo := &mockSequenceRepo{currentNumber: 0}
	docRepo := &mockDocumentRepo{existingNumbers: make(map[string]bool)}

	svc := NewSequenceService(seqRepo, docRepo, txManager)

	t.Run("Генерация_прихода", func(t *testing.T) {
		number, err := svc.GenerateNextDocumentNumber("INCOME")
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(number, "ПР-"))
		assert.Equal(t, "ПР-000001", number)
	})

	t.Run("Генерация_расхода", func(t *testing.T) {
		number, err := svc.GenerateNextDocumentNumber("OUTCOME")
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(number, "РН-"))
		assert.Equal(t, "РН-000002", number)
	})

	t.Run("Обработка_коллизий", func(t *testing.T) {
		seqRepo.currentNumber = 0
		docRepo.existingNumbers["ПМ-000001"] = true

		number, err := svc.GenerateNextDocumentNumber("TRANSFER")
		assert.NoError(t, err)
		assert.Equal(t, "ПМ-000002", number)
	})
}
