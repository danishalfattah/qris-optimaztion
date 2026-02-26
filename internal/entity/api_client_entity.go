package entity

import "time"

type ApiClient struct {
	ClientID     string    `gorm:"column:client_id;primaryKey"`
	ClientSecret string    `gorm:"column:client_secret"`
	Status       string    `gorm:"column:status;default:ACTIVE"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (a *ApiClient) TableName() string {
	return "api_clients"
}
