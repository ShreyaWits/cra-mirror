package user

import (
	"encoding/json"
	"encryption_microservice/pkg/http"
	"fmt"
	"os"
	"sync"
)

const userDataFile = "users.json"

var mu sync.Mutex

type MockFileUserService struct{}

func NewMockUserService(httpClient http.HttpClient, baseURL string) UserService {
	return &MockFileUserService{}
}

// Ensure users.json exists, create if missing
func ensureFileExists() error {
	mu.Lock()
	defer mu.Unlock()

	if _, err := os.Stat(userDataFile); os.IsNotExist(err) {
		emptyData := []User{}
		data, err := json.MarshalIndent(emptyData, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(userDataFile, data, 0644)
	}
	return nil
}

// Load all users from users.json
func loadUsers() ([]User, error) {
	if err := ensureFileExists(); err != nil {
		return nil, err
	}

	file, err := os.ReadFile(userDataFile)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(file, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Save all users to users.json
func saveUsers(users []User) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(userDataFile, data, 0644)
}

// GetUserData retrieves a user by "token" (simulated as ID). If not found, it creates one.
func (s *MockFileUserService) GetUserData(token string) (*User, error) {
	users, err := loadUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == token {
			return &user, nil
		}
	}

	// User not found, create a default one
	newUser := &User{
		ID:          token,
		Name:        "Default Name",
		Email:       "default@example.com",
		EDEKPrivate: "",
		EDEKPublic:  "",
	}

	users = append(users, *newUser)
	if err := saveUsers(users); err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	return newUser, nil
}

// CreateUser adds a new user
func (s *MockFileUserService) CreateUser(token string, newUser *User) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.ID == newUser.ID {
			return fmt.Errorf("user with ID %s already exists", newUser.ID)
		}
	}

	users = append(users, *newUser)
	return saveUsers(users)
}

// DeleteUser removes a user by ID
func (s *MockFileUserService) DeleteUser(token, userID string) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}

	newUsers := make([]User, 0, len(users))
	found := false
	for _, user := range users {
		if user.ID != userID {
			newUsers = append(newUsers, user)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return saveUsers(newUsers)
}

// UpdateUser updates the EDEKPrivate, EDEKPublic, and Name of the user by ID
func (s *MockFileUserService) UpdateUser(userID, edekPrivate, edekPublic string) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}

	var updated bool
	for i, user := range users {
		if user.ID == userID {
			users[i].EDEKPrivate = edekPrivate
			users[i].EDEKPublic = edekPublic
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return saveUsers(users)
}
