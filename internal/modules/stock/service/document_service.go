package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type DocumentService interface {
	Post(id uint) error
	Cancel(id uint) error

	Create(doc *stock.Document) (*stock.Document, error)
	GetByID(id uint) (*stock.Document, error)
	List() ([]stock.Document, error)
	ListByStatus(status string) ([]stock.Document, error)
	Update(doc *stock.Document) (*stock.Document, error)
	Delete(id uint) error
}

type documentService struct {
	repo         repository.DocumentRepository
	historyRepo  repository.DocumentHistoryRepository
	inventory    InventoryService
	priceService PriceService
	sequenceSvc  SequenceService
	tx           repository.TxManager
}

func NewDocumentService(
	r repository.DocumentRepository,
	h repository.DocumentHistoryRepository,
	inv InventoryService,
	priceSvc PriceService,
	seqSvc SequenceService,
	tx repository.TxManager,
) DocumentService {
	return &documentService{
		repo:         r,
		historyRepo:  h,
		inventory:    inv,
		priceService: priceSvc,
		sequenceSvc:  seqSvc,
		tx:           tx,
	}
}

func (s *documentService) Post(id uint) error {
	return s.tx.DoInTx(func(tx *gorm.DB) error {
		doc, err := s.repo.GetByIDWithTx(tx, id)
		if err != nil {
			return err
		}
		if doc == nil {
			return errors.New("document not found")
		}
		if doc.Status == "posted" {
			return errors.New("document already posted")
		}
		if doc.Status == "canceled" {
			return errors.New("cannot post canceled document")
		}

		switch toUpper(doc.Type) {
		case "INCOME", "OUTCOME", "ORDER", "TRANSFER", "INVENTORY":
			if err := s.inventory.ProcessDocumentWithTx(tx, doc); err != nil {
				return fmt.Errorf("inventory processing failed: %w", err)
			}
		case "PRICE_UPDATE":
			if err := s.priceService.UpdatePricesFromDocumentWithTx(tx, doc); err != nil {
				return fmt.Errorf("price update processing failed: %w", err)
			}
		default:
			return fmt.Errorf("unknown document type to post: '%s'", doc.Type)
		}

		now := time.Now()
		doc.Status = "posted"
		doc.PostedAt = &now
		if _, err := s.repo.UpdateWithTx(tx, doc); err != nil {
			return err
		}

		h := &stock.DocumentHistory{
			DocumentID: doc.ID,
			Action:     "posted",
			CreatedAt:  now,
			CreatedBy:  doc.CreatedBy,
		}
		return s.historyRepo.CreateWithTx(tx, h)
	})
}

func (s *documentService) Cancel(id uint) error {
	return s.tx.DoInTx(func(tx *gorm.DB) error {
		doc, err := s.repo.GetByIDWithTx(tx, id)
		if err != nil {
			return err
		}
		if doc == nil {
			return errors.New("document not found")
		}
		if doc.Status != "posted" {
			return errors.New("only posted documents can be canceled")
		}

		switch toUpper(doc.Type) {
		case "INCOME", "OUTCOME", "ORDER", "TRANSFER", "INVENTORY":
			if err := s.inventory.RevertDocumentWithTx(tx, doc); err != nil {
				return err
			}
		case "PRICE_UPDATE":
			return errors.New("cancellation for 'PRICE_UPDATE' is not yet implemented")
		default:
			return fmt.Errorf("unknown document type to cancel: '%s'", doc.Type)
		}

		doc.Status = "canceled"
		if _, err := s.repo.UpdateWithTx(tx, doc); err != nil {
			return err
		}

		h := &stock.DocumentHistory{
			DocumentID: doc.ID,
			Action:     "canceled",
			CreatedAt:  time.Now(),
			CreatedBy:  doc.CreatedBy,
		}
		return s.historyRepo.CreateWithTx(tx, h)
	})
}

func (s *documentService) Create(doc *stock.Document) (*stock.Document, error) {
	if doc.Status == "" {
		doc.Status = "draft"
	}

	newNumber, err := s.sequenceSvc.GenerateNextDocumentNumber(doc.Type)
	if err != nil {
		return nil, fmt.Errorf("could not generate document number: %w", err)
	}
	doc.Number = newNumber

	return s.repo.Create(doc)
}

func (s *documentService) GetByID(id uint) (*stock.Document, error) {
	return s.repo.GetByID(id)
}

func (s *documentService) List() ([]stock.Document, error) {
	return s.repo.List()
}

func (s *documentService) ListByStatus(status string) ([]stock.Document, error) {
	return s.repo.ListByStatus(status)
}

func (s *documentService) Update(doc *stock.Document) (*stock.Document, error) {
	return s.repo.Update(doc)
}

func (s *documentService) Delete(id uint) error {
	return s.repo.Delete(id)
}
