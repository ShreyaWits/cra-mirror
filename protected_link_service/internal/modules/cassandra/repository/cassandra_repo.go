// Package repository defines the repository interfaces for domain logic.
package repository

import (
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"

	"github.com/gocql/gocql"
)

// ICassandraRepository defines the interface for Cassandra database operations.
type ICassandraRepository interface {
	// GetData fetches data by key.
	GetDataByID(id string) (*apiDtos.GenerateUrlRequest, error)

	DeleteById(id string) error
	// SaveData saves data by key.
	SaveData(dto *apiDtos.GenerateUrlRequest) (gocql.UUID, error)
}
