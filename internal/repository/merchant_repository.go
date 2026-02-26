package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MerchantRepository struct {
	Repository[entity.Merchant]
	Log *logrus.Logger
}

func NewMerchantRepository(log *logrus.Logger) *MerchantRepository {
	return &MerchantRepository{
		Log: log,
	}
}

func (r *MerchantRepository) FindByMerchantID(db *gorm.DB, merchant *entity.Merchant, merchantID string) error {
	return db.Where("merchant_id = ? AND is_active = ?", merchantID, true).Take(merchant).Error
}

func (r *MerchantRepository) FindAll(db *gorm.DB) ([]entity.Merchant, error) {
	var merchants []entity.Merchant
	err := db.Where("is_active = ?", true).Find(&merchants).Error
	return merchants, err
}
