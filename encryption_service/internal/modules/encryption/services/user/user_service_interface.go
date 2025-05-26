package user

import (
	"context"
	"encryption_microservice/pkg/errors"
)

// UserServiceInterface defines the methods for interacting with the user service
type UserService interface {
	// GetUserData fetches user data by user ID
	GetUserData(ctx context.Context, token string) (*User, *errors.CustomError)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, token string, user *User) *errors.CustomError

	// DeleteUser deletes a user by user ID
	DeleteUser(ctx context.Context, token string, userID string) *errors.CustomError

	UpdateUser(ctx context.Context, userID, edekPrivate, edekPublic string) *errors.CustomError
}
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	EDEKPrivate string `json:"edekPrivate"`
	EDEKPublic  string `json:"edekPublic"`
}
