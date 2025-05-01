package models

type AuthRequestData struct {
	AuthProvider string   `json:"auth_provider" validate:"required" example:"google"`
	Permissions  []string `json:"permissions" validate:"required" example:"read,write"`
}
