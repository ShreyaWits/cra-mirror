package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Template struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex" json:"name"`
	Content   string         `json:"content"`
	Channel   string         `gorm:"index" json:"channel"`
	Language  string         `gorm:"index" json:"language"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	Version   int            `gorm:"default:1" json:"version"`
	CreatedBy string         `json:"created_by"`
	UpdatedBy string         `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Hook to generate UUID before create
func (t *Template) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
