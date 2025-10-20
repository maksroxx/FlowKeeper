package repository

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type DocumentRepository interface {
	Create(doc *stock.Document) (*stock.Document, error)
	List() ([]stock.Document, error)
	Update(doc *stock.Document) (*stock.Document, error)
	Delete(id uint) error

	GetByID(id uint) (*stock.Document, error)
	GetByNumber(number string) (*stock.Document, error)
	ListByStatus(status string) ([]stock.Document, error)

	UpdateWithTx(tx *gorm.DB, doc *stock.Document) (*stock.Document, error)
	CreateWithTx(tx *gorm.DB, doc *stock.Document) (*stock.Document, error)
	GetByIDWithTx(tx *gorm.DB, id uint) (*stock.Document, error)

	GetByIDs(ids []uint) ([]stock.Document, error)

	Search(filter stock.DocumentFilter) ([]stock.Document, error)
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

func (r *documentRepo) GetByIDWithTx(tx *gorm.DB, id uint) (*stock.Document, error) {
	var d stock.Document
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Items").
		First(&d, id).Error

	if err != nil {
		return nil, err
	}
	return &d, nil
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

func (r *documentRepo) ListByStatus(status string) ([]stock.Document, error) {
	var docs []stock.Document
	if err := r.db.Preload("Items").Where("status = ?", status).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
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

func (r *documentRepo) GetByIDs(ids []uint) ([]stock.Document, error) {
	var documents []stock.Document
	if len(ids) == 0 {
		return documents, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&documents).Error
	return documents, err
}

func (r *documentRepo) Search(filter stock.DocumentFilter) ([]stock.Document, error) {
	var docs []stock.Document
	query := r.db.Model(&stock.Document{}).Preload("Items")

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if len(filter.Types) > 0 {
		query = query.Where("type IN ?", filter.Types)
	}

	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}

	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}

	if filter.Search != nil {
		searchPattern := "%" + strings.ToLower(*filter.Search) + "%"
		query = query.Where("LOWER(number) LIKE ? OR LOWER(comment) LIKE ?", searchPattern, searchPattern)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("created_at desc").Find(&docs).Error
	return docs, err
}
