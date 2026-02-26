package model

// InquiryResponse represents the QRIS inquiry result
type InquiryResponse struct {
	MerchantID   string  `json:"merchant_id"`
	MerchantName string  `json:"merchant_name"`
	TerminalID   string  `json:"terminal_id"`
	City         string  `json:"city"`
	FixedAmount  float64 `json:"fixed_amount"`
	InquiryID    string  `json:"inquiry_id"`
}

// PaymentRequest represents the QRIS payment request body
type PaymentRequest struct {
	InquiryID     string  `json:"inquiry_id" validate:"required"`
	UserID        string  `json:"user_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" validate:"required"`
	Pincode       string  `json:"pincode" validate:"required"`
}

// PaymentResponse represents the QRIS payment result
type PaymentResponse struct {
	Status              string `json:"status"`
	TransactionID       string `json:"transaction_id"`
	Message             string `json:"message"`
	EstimatedCompletion string `json:"estimated_completion"`
}
