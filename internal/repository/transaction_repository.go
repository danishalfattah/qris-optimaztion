package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	Repository[entity.Transaction]
	Log *logrus.Logger
}

func NewTransactionRepository(log *logrus.Logger) *TransactionRepository {
	return &TransactionRepository{
		Log: log,
	}
}

func (r *TransactionRepository) FindByTransactionID(db *gorm.DB, transaction *entity.Transaction, transactionID string) error {
	return db.Where("transaction_id = ?", transactionID).Take(transaction).Error
}

func (r *TransactionRepository) UpdateStatus(db *gorm.DB, transactionID string, status string) error {
	return db.Model(&entity.Transaction{}).
		Where("transaction_id = ?", transactionID).
		Update("status", status).Error
}
