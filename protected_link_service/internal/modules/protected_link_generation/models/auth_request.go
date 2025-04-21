package models

type AuthRequestData struct {
	UserID       string   `json:"user_id" validate:"required" example:"1234"`
	AuthProvider string   `json:"auth_provider" validate:"required" example:"google"`
	Permissions  []string `json:"permissions" validate:"required" example:"read,write"`
}
