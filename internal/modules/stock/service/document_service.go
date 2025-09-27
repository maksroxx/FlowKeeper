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
	Create(doc *stock.Document) (*stock.Document, error)
	List() ([]stock.Document, error)
	GetByID(id uint) (*stock.Document, error)
	Update(doc *stock.Document) (*stock.Document, error)
	Delete(id uint) error

	Post(id uint) error
	Cancel(id uint) error
}

type documentService struct {
	repo        repository.DocumentRepository
	historyRepo repository.DocumentHistoryRepository
	inventory   InventoryService
	tx          repository.TxManager
}

func NewDocumentService(r repository.DocumentRepository, h repository.DocumentHistoryRepository, inv InventoryService, tx repository.TxManager) DocumentService {
	return &documentService{
		repo:        r,
		historyRepo: h,
		inventory:   inv,
		tx:          tx,
	}
}

func (s *documentService) Create(doc *stock.Document) (*stock.Document, error) {
	if doc.Status == "" {
		doc.Status = "draft"
	}
	if doc.Number != "" {
		existing, _ := s.repo.GetByNumber(doc.Number)
		if existing != nil {
			return nil, fmt.Errorf("document with number %s already exists", doc.Number)
		}
	}
	return s.repo.Create(doc)
}

func (s *documentService) List() ([]stock.Document, error) {
	return s.repo.List()
}

func (s *documentService) GetByID(id uint) (*stock.Document, error) {
	return s.repo.GetByID(id)
}

func (s *documentService) Update(doc *stock.Document) (*stock.Document, error) {
	return s.repo.Update(doc)
}

func (s *documentService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *documentService) Post(id uint) error {
	return s.tx.DoInTx(func(tx *gorm.DB) error {
		doc, err := s.repo.GetByID(id)
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

		if err := s.inventory.ProcessDocumentWithTx(tx, doc); err != nil {
			return err
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
			Comment:    doc.Comment,
		}
		if err := s.historyRepo.CreateWithTx(tx, h); err != nil {
			return err
		}

		return nil
	})
}

func (s *documentService) Cancel(id uint) error {
	return s.tx.DoInTx(func(tx *gorm.DB) error {
		doc, err := s.repo.GetByID(id)
		if err != nil {
			return err
		}
		if doc == nil {
			return errors.New("document not found")
		}
		if doc.Status != "posted" {
			return errors.New("only posted documents can be canceled")
		}

		if err := s.inventory.RevertDocumentWithTx(tx, doc); err != nil {
			return err
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
			Comment:    doc.Comment,
		}
		if err := s.historyRepo.CreateWithTx(tx, h); err != nil {
			return err
		}

		return nil
	})
}
