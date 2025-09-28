package repository

import (
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"gorm.io/gorm"
)

type ReservationRepository interface {
	GetReservationWithTx(tx *gorm.DB, warehouseID, variantID uint) (*models.StockReservation, error)
	SaveReservationWithTx(tx *gorm.DB, r *models.StockReservation) error
}

type reservationRepo struct{ db *gorm.DB }

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepo{db: db}
}

func (r *reservationRepo) GetReservationWithTx(tx *gorm.DB, warehouseID, variantID uint) (*models.StockReservation, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var res models.StockReservation
	err := db.Where("warehouse_id = ? AND variant_id = ?", warehouseID, variantID).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &res, nil
}

func (r *reservationRepo) SaveReservationWithTx(tx *gorm.DB, res *models.StockReservation) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(res).Error
}
