package models

import (
	"time"
)

type BaseModel struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
