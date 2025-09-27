package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
)

type PriceRepository interface {
	UpsertPrices(tx *gorm.DB, prices []models.ItemPrice) error
	GetPrice(itemID, priceTypeID uint) (*models.ItemPrice, error)
}

type priceRepo struct {
	db *gorm.DB
}

func NewPriceRepository(db *gorm.DB) PriceRepository {
	return &priceRepo{db: db}
}

func (r *priceRepo) UpsertPrices(tx *gorm.DB, prices []models.ItemPrice) error {
	if len(prices) == 0 {
		return nil
	}

	now := time.Now()
	for i := range prices {
		prices[i].UpdatedAt = now
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "item_id"}, {Name: "price_type_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "currency", "updated_at"}),
	}).Create(&prices).Error
}

func (r *priceRepo) GetPrice(itemID, priceTypeID uint) (*models.ItemPrice, error) {
	var price models.ItemPrice
	err := r.db.Where("item_id = ? AND price_type_id = ?", itemID, priceTypeID).First(&price).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &price, nil
}
