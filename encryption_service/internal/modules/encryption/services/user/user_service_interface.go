package user

// UserServiceInterface defines the methods for interacting with the user service
type UserService interface {
	// GetUserData fetches user data by user ID
	GetUserData(token string) (*User, error)

	// CreateUser creates a new user
	CreateUser(token string, user *User) error

	// DeleteUser deletes a user by user ID
	DeleteUser(token string, userID string) error

	UpdateUser(userID, edekPrivate, edekPublic string) error
}
