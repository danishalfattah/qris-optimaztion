package entity

type Merchant struct {
	MerchantID   string `gorm:"column:merchant_id;primaryKey"`
	MerchantName string `gorm:"column:merchant_name"`
	MCC          string `gorm:"column:mcc"`
	City         string `gorm:"column:city"`
	IsActive     bool   `gorm:"column:is_active;default:true"`
}

func (m *Merchant) TableName() string {
	return "merchants"
}
