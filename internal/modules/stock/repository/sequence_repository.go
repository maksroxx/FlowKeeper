package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type SequenceRepository interface {
	GetNext(tx *gorm.DB, id string) (*models.DocumentSequence, error)
}

type sequenceRepo struct{}

func NewSequenceRepository() SequenceRepository {
	return &sequenceRepo{}
}

func (r *sequenceRepo) GetNext(tx *gorm.DB, id string) (*models.DocumentSequence, error) {
	var seq models.DocumentSequence
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seq, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			seq = models.DocumentSequence{
				ID:         id,
				LastNumber: 0,
			}
		} else {
			return nil, err
		}
	}

	seq.LastNumber++

	if err := tx.Save(&seq).Error; err != nil {
		return nil, err
	}

	return &seq, nil
}
