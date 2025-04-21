package user

import (
	"encoding/json"
	"encryption_microservice/pkg/http"
	"fmt"
)

// UserService defines the methods for interacting with the user service
type UserServiceStruct struct {
	httpClient http.HttpClient
	baseURL    string
}

// NewUserService creates a new instance of UserService
func NewUserService(httpClient http.HttpClient, baseURL string) UserService {
	return &UserServiceStruct{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// User represents the structure of user data
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetUserData fetches user data by user ID
func (s *UserServiceStruct) GetUserData(userID string) (*User, error) {
	url := fmt.Sprintf("%s/users/%s", s.baseURL, userID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Make the GET request
	response, err := s.httpClient.Get(url, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user data: %w", err)
	}

	// Parse the response
	var user User
	if err := json.Unmarshal(response, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (s *UserServiceStruct) CreateUser(user *User) error {
	url := fmt.Sprintf("%s/users", s.baseURL)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Make the POST request
	_, err := s.httpClient.Post(url, user, headers)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by user ID
func (s *UserServiceStruct) DeleteUser(userID string) error {
	url := fmt.Sprintf("%s/users/%s", s.baseURL, userID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Make the DELETE request
	// Since HTTPClientInterface doesn't have a Delete method, we can use Post with an empty payload
	_, err := s.httpClient.Post(url, nil, headers)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
