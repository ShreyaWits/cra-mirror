package models

type AuthRequestData struct {
    AuthProvider string   `json:"auth_provider"`
    Permissions  string   `json:"permissions"` // Change to string
}
