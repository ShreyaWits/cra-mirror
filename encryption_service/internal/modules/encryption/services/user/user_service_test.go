package user_test

import (
	"encryption_microservice/internal/modules/encryption/services/user"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFile = "users.json"

func setup(t *testing.T) {
	// Clean test file before each test
	err := os.Remove(testFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to cleanup: %v", err)
	}
}

func TestGetUserData_CreatesUserIfNotExist(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}

	token := "test123"
	u, err := svc.GetUserData(token)

	assert.Nil(t, err)
	assert.Equal(t, token, u.ID)
	assert.Equal(t, "Default Name", u.Name)
	assert.Equal(t, "default@example.com", u.Email)
}

func TestGetUserData_ReturnsExistingUser(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	token := "existingUser"

	// Create the user manually
	_ = svc.CreateUser(token, &user.User{
		ID:    token,
		Name:  "Alice",
		Email: "alice@example.com",
	})

	// Now retrieve it
	u, err := svc.GetUserData(token)
	assert.Nil(t, err)
	assert.Equal(t, "Alice", u.Name)
	assert.Equal(t, "alice@example.com", u.Email)
}

func TestCreateUser_FailsIfExists(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	token := "dupeUser"

	err := svc.CreateUser(token, &user.User{
		ID:    token,
		Name:  "Bob",
		Email: "bob@example.com",
	})
	assert.Nil(t, err)

	err = svc.CreateUser(token, &user.User{
		ID:    token,
		Name:  "Bob Again",
		Email: "bob2@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteUser_RemovesUser(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	token := "deleteUser"

	_ = svc.CreateUser(token, &user.User{
		ID:    token,
		Name:  "To Delete",
		Email: "delete@example.com",
	})

	err := svc.DeleteUser(token, token)
	assert.Nil(t, err)

	// Try to get it again
	_, err = svc.GetUserData(token)
	assert.Nil(t, err) // Should recreate
}

func TestUpdateUser_UpdatesFields(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	token := "updateUser"

	_ = svc.CreateUser(token, &user.User{
		ID:          token,
		Name:        "Old",
		Email:       "old@example.com",
		EDEKPublic:  "",
		EDEKPrivate: "",
	})

	err := svc.UpdateUser(token, "private-edek", "public-edek")
	assert.Nil(t, err)

	u, err2 := svc.GetUserData(token)
	assert.Nil(t, err2)
	assert.Equal(t, "private-edek", u.EDEKPrivate)
	assert.Equal(t, "public-edek", u.EDEKPublic)
}
