package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AccountRepository struct {
	Repository[entity.Account]
	Log *logrus.Logger
}

func NewAccountRepository(log *logrus.Logger) *AccountRepository {
	return &AccountRepository{
		Log: log,
	}
}

func (r *AccountRepository) FindByAccountID(db *gorm.DB, account *entity.Account, accountID string) error {
	return db.Where("account_id = ?", accountID).Take(account).Error
}

// DeductBalance uses optimistic locking to prevent double-spend
func (r *AccountRepository) DeductBalance(db *gorm.DB, accountID string, amount float64, expectedVersion int) error {
	result := db.Model(&entity.Account{}).
		Where("account_id = ? AND version = ? AND balance >= ?", accountID, expectedVersion, amount).
		Updates(map[string]interface{}{
			"balance": gorm.Expr("balance - ?", amount),
			"version": gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
