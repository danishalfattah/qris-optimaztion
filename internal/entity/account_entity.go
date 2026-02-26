package entity

type Account struct {
	AccountID string  `gorm:"column:account_id;primaryKey"`
	Balance   float64 `gorm:"column:balance;type:decimal(18,2);default:0"`
	Currency  string  `gorm:"column:currency;default:IDR"`
	PinHash   string  `gorm:"column:pin_hash"`
	Version   int     `gorm:"column:version;default:0"`
}

func (a *Account) TableName() string {
	return "accounts"
}
