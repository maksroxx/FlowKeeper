package repository

import (
	"gorm.io/gorm"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type DocumentRepository interface {
	Create(doc *stock.Document) (*stock.Document, error)
	List() ([]stock.Document, error)
	Update(doc *stock.Document) (*stock.Document, error)
	Delete(id uint) error

	GetByID(id uint) (*stock.Document, error)
	GetByNumber(number string) (*stock.Document, error)

	UpdateWithTx(tx *gorm.DB, doc *stock.Document) (*stock.Document, error)
	CreateWithTx(tx *gorm.DB, doc *stock.Document) (*stock.Document, error)
}

type documentRepo struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) Create(doc *stock.Document) (*stock.Document, error) {
	if err := r.db.Create(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *documentRepo) CreateWithTx(tx *gorm.DB, doc *stock.Document) (*stock.Document, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Create(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *documentRepo) GetByID(id uint) (*stock.Document, error) {
	var d stock.Document
	if err := r.db.Preload("Items").First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *documentRepo) GetByNumber(number string) (*stock.Document, error) {
	var d stock.Document
	if err := r.db.Where("number = ?", number).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *documentRepo) List() ([]stock.Document, error) {
	var docs []stock.Document
	if err := r.db.Preload("Items").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *documentRepo) Update(doc *stock.Document) (*stock.Document, error) {
	if err := r.db.Save(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *documentRepo) UpdateWithTx(tx *gorm.DB, doc *stock.Document) (*stock.Document, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Save(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *documentRepo) Delete(id uint) error {
	return r.db.Delete(&stock.Document{}, id).Error
}
