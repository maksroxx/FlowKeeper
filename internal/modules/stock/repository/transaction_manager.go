package repository

import "gorm.io/gorm"

type TxManager interface {
	DoInTx(func(tx *gorm.DB) error) error
}

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) TxManager {
	return &txManager{db: db}
}

func (m *txManager) DoInTx(fn func(tx *gorm.DB) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
