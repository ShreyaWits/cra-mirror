package user

// UserServiceInterface defines the methods for interacting with the user service
type UserService interface {
	// GetUserData fetches user data by user ID
	GetUserData(userID string) (*User, error)

	// CreateUser creates a new user
	CreateUser(user *User) error

	// DeleteUser deletes a user by user ID
	DeleteUser(userID string) error
}
