package model

// TransactionStatusResponse represents the transaction status result
type TransactionStatusResponse struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	FinalBalance  float64 `json:"final_balance"`
	Timestamp     string  `json:"timestamp"`
}
