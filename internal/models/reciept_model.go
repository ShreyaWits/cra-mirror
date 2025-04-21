package models

type ClaimRequest struct {
	PRAN            string `json:"pran" validate:"required,len=12,numeric"`
	DateOfClaim     string `json:"date_of_claim" validate:"required,datetime=2006-01-02"`
	TransactionType string `json:"transaction_type" validate:"required"`
}

type ClaimResponse struct {
	ClaimId    string `json:"claim_id"`
}