package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ApiClientRepository struct {
	Repository[entity.ApiClient]
	Log *logrus.Logger
}

func NewApiClientRepository(log *logrus.Logger) *ApiClientRepository {
	return &ApiClientRepository{
		Log: log,
	}
}

func (r *ApiClientRepository) FindByClientID(db *gorm.DB, client *entity.ApiClient, clientID string) error {
	return db.Where("client_id = ? AND status = ?", clientID, "ACTIVE").Take(client).Error
}
