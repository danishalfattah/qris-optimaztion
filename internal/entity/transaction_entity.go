package entity

import "time"

type Transaction struct {
	TransactionID string    `gorm:"column:transaction_id;primaryKey;type:uuid;default:gen_random_uuid()"`
	TraceID       string    `gorm:"column:trace_id;index"`
	AccountID     string    `gorm:"column:account_id"`
	MerchantID    string    `gorm:"column:merchant_id"`
	Amount        float64   `gorm:"column:amount;type:decimal(18,2)"`
	Status        string    `gorm:"column:status;default:PENDING"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	Account       Account   `gorm:"foreignKey:AccountID;references:AccountID"`
	Merchant      Merchant  `gorm:"foreignKey:MerchantID;references:MerchantID"`
}

func (t *Transaction) TableName() string {
	return "transactions"
}
