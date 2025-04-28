package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Config struct {
	ServiceName string                 `json:"service_name"`
	Environment string                 `json:"environment"`
	ConfigData  map[string]interface{} `json:"config_data"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ConfigMetadata struct {
	LastModifiedBy string    `json:"last_modified_by"`
	ChangeHistory  []string  `json:"change_history"`
	LastModifiedAt time.Time `json:"last_modified_at"`
}

type Admin struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserName    string    `gorm:"not null"`
	Password    string    `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (n *Admin) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return
}
