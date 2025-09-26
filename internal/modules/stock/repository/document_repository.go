package repository

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	Create(doc *stock.Document) (*stock.Document, error)
	GetByID(id uint) (*stock.Document, error)
	List() ([]stock.Document, error)
	Update(doc *stock.Document) (*stock.Document, error)
	Delete(id uint) error
}

type documentRepository struct{ db *gorm.DB }

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(doc *stock.Document) (*stock.Document, error) {
	err := r.db.Create(doc).Error
	return doc, err
}

func (r *documentRepository) GetByID(id uint) (*stock.Document, error) {
	var doc stock.Document
	if err := r.db.Preload("Items").First(&doc, id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) List() ([]stock.Document, error) {
	var docs []stock.Document
	err := r.db.Preload("Items").Find(&docs).Error
	return docs, err
}

func (r *documentRepository) Update(doc *stock.Document) (*stock.Document, error) {
	err := r.db.Save(doc).Error
	return doc, err
}

func (r *documentRepository) Delete(id uint) error {
	return r.db.Delete(&stock.Document{}, id).Error
}
