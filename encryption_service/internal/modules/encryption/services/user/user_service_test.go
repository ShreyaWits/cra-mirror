package user_test

import (
	"context"
	"encryption_microservice/internal/modules/encryption/services/user"
	"encryption_microservice/pkg/errors"
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
	ctx := context.Background()

	token := "test123"
	u, err := svc.GetUserData(ctx, token)

	assert.Nil(t, err)
	assert.Equal(t, token, u.ID)
	assert.Equal(t, "Default Name", u.Name)
	assert.Equal(t, "default@example.com", u.Email)
}

func TestGetUserData_ReturnsExistingUser(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	token := "existingUser"

	// Create the user manually
	_ = svc.CreateUser(ctx, token, &user.User{
		ID:    token,
		Name:  "Alice",
		Email: "alice@example.com",
	})

	// Now retrieve it
	u, err := svc.GetUserData(ctx, token)
	assert.Nil(t, err)
	assert.Equal(t, "Alice", u.Name)
	assert.Equal(t, "alice@example.com", u.Email)
}

func TestCreateUser_FailsIfExists(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	token := "dupeUser"

	err := svc.CreateUser(ctx, token, &user.User{
		ID:    token,
		Name:  "Bob",
		Email: "bob@example.com",
	})
	assert.Nil(t, err)

	err = svc.CreateUser(ctx, token, &user.User{
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
	ctx := context.Background()
	token := "deleteUser"

	_ = svc.CreateUser(ctx, token, &user.User{
		ID:    token,
		Name:  "To Delete",
		Email: "delete@example.com",
	})

	err := svc.DeleteUser(ctx, token, token)
	assert.Nil(t, err)

	// Try to get it again
	_, err = svc.GetUserData(ctx, token)
	assert.Nil(t, err) // Should recreate
}

func TestUpdateUser_UpdatesFields(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	token := "updateUser"

	_ = svc.CreateUser(ctx, token, &user.User{
		ID:          token,
		Name:        "Old",
		Email:       "old@example.com",
		EDEKPublic:  "",
		EDEKPrivate: "",
	})

	err := svc.UpdateUser(ctx, token, "private-edek", "public-edek")
	assert.Nil(t, err)

	u, err2 := svc.GetUserData(ctx, token)
	assert.Nil(t, err2)
	assert.Equal(t, "private-edek", u.EDEKPrivate)
	assert.Equal(t, "public-edek", u.EDEKPublic)
}

func TestGetUserData_ErrorIfTokenEmpty(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	u, err := svc.GetUserData(ctx, "")
	assert.Nil(t, u)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errors.USRErrTokenRequired)
}

func TestCreateUser_Succeeds(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	token := "newUser"
	newUser := &user.User{
		ID:    token,
		Name:  "New Name",
		Email: "new@example.com",
	}
	err := svc.CreateUser(ctx, token, newUser)
	assert.Nil(t, err)
	u, err2 := svc.GetUserData(ctx, token)
	assert.Nil(t, err2)
	assert.Equal(t, "New Name", u.Name)
	assert.Equal(t, "new@example.com", u.Email)
}

func TestDeleteUser_ErrorIfNotFound(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	err := svc.DeleteUser(ctx, "token", "nonexistent")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errors.USRErrDeleteUser)
}

func TestUpdateUser_ErrorIfNotFound(t *testing.T) {
	setup(t)
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	err := svc.UpdateUser(ctx, "nonexistent", "p", "q")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), errors.USRErrUpdateUser)
}

func TestGetUserData_ErrorIfCorruptedFile(t *testing.T) {
	setup(t)
	// Write invalid JSON to users.json
	err := os.WriteFile(testFile, []byte("not a valid json"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid data: %v", err)
	}
	svc := &user.MockFileUserService{}
	ctx := context.Background()
	u, err2 := svc.GetUserData(ctx, "token123")
	assert.Nil(t, u)
	assert.NotNil(t, err2)
	assert.Contains(t, err2.Error(), errors.USRErrFetchUserData)
}
