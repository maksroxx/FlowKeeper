package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"gorm.io/gorm"
)

type SequenceService interface {
	GenerateNextDocumentNumber(docType string) (string, error)
}

type sequenceService struct {
	repo repository.SequenceRepository
	tx   repository.TxManager
}

func NewSequenceService(r repository.SequenceRepository, tx repository.TxManager) SequenceService {
	return &sequenceService{repo: r, tx: tx}
}

func (s *sequenceService) GenerateNextDocumentNumber(docType string) (string, error) {
	var finalNumber string
	var err error

	year := time.Now().Year()
	upperType := strings.ToUpper(docType)
	sequenceID := fmt.Sprintf("%s_%d", upperType, year)

	err = s.tx.DoInTx(func(tx *gorm.DB) error {
		seq, txErr := s.repo.GetNext(tx, sequenceID)
		if txErr != nil {
			return txErr
		}

		var prefix string
		switch upperType {
		case "INCOME":
			prefix = "ПР"
		case "OUTCOME":
			prefix = "РН"
		case "TRANSFER":
			prefix = "ПМ"
		case "INVENTORY":
			prefix = "ИН"
		default:
			prefix = "ДОК"
		}

		finalNumber = fmt.Sprintf("%s-%06d", prefix, seq.LastNumber)
		return nil
	})

	return finalNumber, err
}
