package service

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type DocumentService interface {
	Create(doc *stock.Document) (*stock.Document, error)
	List() ([]stock.Document, error)
	GetByID(id uint) (*stock.Document, error)
}

type documentService struct {
	repo repository.DocumentRepository
}

func NewDocumentService(r repository.DocumentRepository) DocumentService {
	return &documentService{repo: r}
}

func (s *documentService) Create(doc *stock.Document) (*stock.Document, error) {
	return s.repo.Create(doc)
}

func (s *documentService) List() ([]stock.Document, error) {
	return s.repo.List()
}

func (s *documentService) Update(i *stock.Document) (*stock.Document, error) { return s.repo.Update(i) }
func (s *documentService) Delete(id uint) error                              { return s.repo.Delete(id) }

func (s *documentService) GetByID(id uint) (*stock.Document, error) {
	return s.repo.GetByID(id)
}
