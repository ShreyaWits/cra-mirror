package authUtil

import (
	"errors"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

func ParseUUID(idStr string) (gocql.UUID, error) {
	if idStr == "" {
		return gocql.UUID{}, errors.New("id cannot be empty")
	}

	parsedUUID, err := uuid.Parse(idStr)
	if err != nil {
		return gocql.UUID{}, err
	}

	gocqlUUID, err := gocql.ParseUUID(parsedUUID.String())
	if err != nil {
		return gocql.UUID{}, err
	}

	return gocqlUUID, nil
}
