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
	repo    repository.SequenceRepository
	docRepo repository.DocumentRepository
	tx      repository.TxManager
}

func NewSequenceService(r repository.SequenceRepository, docRepo repository.DocumentRepository, tx repository.TxManager) SequenceService {
	return &sequenceService{repo: r, docRepo: docRepo, tx: tx}
}

func (s *sequenceService) GenerateNextDocumentNumber(docType string) (string, error) {
	var finalNumber string
	var err error

	year := time.Now().Year()
	upperType := strings.ToUpper(docType)
	sequenceID := fmt.Sprintf("%s_%d", upperType, year)

	err = s.tx.DoInTx(func(tx *gorm.DB) error {
		for {
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
			case "PRICE_UPDATE":
				prefix = "УЦ"
			case "ORDER":
				prefix = "ЗК"
			default:
				prefix = "ДОК"
			}

			candidateNumber := fmt.Sprintf("%s-%06d", prefix, seq.LastNumber)
			existingDoc, _ := s.docRepo.GetByNumber(candidateNumber)

			if existingDoc == nil {
				finalNumber = candidateNumber
				return nil
			}
		}
	})

	return finalNumber, err
}
