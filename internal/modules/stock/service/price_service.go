package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"gorm.io/gorm"
)

type PriceService interface {
	UpdatePricesFromDocumentWithTx(tx *gorm.DB, doc *models.Document) error
	GetPrice(itemID, priceTypeID uint) (*models.ItemPrice, error)
}

type priceService struct {
	repo repository.PriceRepository
}

func NewPriceService(r repository.PriceRepository) PriceService {
	return &priceService{repo: r}
}

func (s *priceService) UpdatePricesFromDocumentWithTx(tx *gorm.DB, doc *models.Document) error {
	if doc == nil {
		return errors.New("document is nil")
	}
	if doc.PriceTypeID == nil {
		return errors.New("priceTypeID must be set in the header for a PRICE_UPDATE document")
	}
	if len(doc.Items) == 0 {
		return nil
	}

	var pricesToUpdate []models.ItemPrice
	for _, item := range doc.Items {
		if item.Price == nil {
			return fmt.Errorf("price is not set for item ID %d in document", item.ItemID)
		}

		priceRecord := models.ItemPrice{
			ItemID:      item.ItemID,
			PriceTypeID: *doc.PriceTypeID,
			Price:       *item.Price,
			Currency:    "RUB",
			UpdatedAt:   time.Now(),
		}
		pricesToUpdate = append(pricesToUpdate, priceRecord)
	}

	return s.repo.UpsertPrices(tx, pricesToUpdate)
}

func (s *priceService) GetPrice(itemID, priceTypeID uint) (*models.ItemPrice, error) {
	return s.repo.GetPrice(itemID, priceTypeID)
}
