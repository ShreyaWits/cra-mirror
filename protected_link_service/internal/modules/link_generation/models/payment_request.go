package models

type PaymentRequestData struct {
	TransactionID string  `json:"transaction_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required"`
	Currency      string  `json:"currency" validate:"required"`
}
