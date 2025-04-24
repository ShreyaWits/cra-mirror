package user

import "encryption_microservice/pkg/errors"

// UserServiceInterface defines the methods for interacting with the user service
type UserService interface {
	// GetUserData fetches user data by user ID
	GetUserData(token string) (*User, *errors.CustomError)

	// CreateUser creates a new user
	CreateUser(token string, user *User) *errors.CustomError

	// DeleteUser deletes a user by user ID
	DeleteUser(token string, userID string) *errors.CustomError

	UpdateUser(userID, edekPrivate, edekPublic string) *errors.CustomError
}
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	EDEKPrivate string `json:"edekPrivate"`
	EDEKPublic  string `json:"edekPublic"`
}
