package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	BaseModel

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (u *User) TableName() string {
	return "users"
}

// BeforeCreate GORM hook to set the UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = fmt.Sprintf("user-%s", uuid.New().String())
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	return nil
}

// BeforeUpdate GORM hook to set the UUID
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
