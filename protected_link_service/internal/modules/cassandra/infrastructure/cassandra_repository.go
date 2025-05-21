package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	authUtils "protected_link/internal/modules/authentication/utils"
	repository "protected_link/internal/modules/cassandra/repository"

	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/pkg/cassandra"

	"github.com/gocql/gocql"
)

// CassandraClient is a placeholder for an actual Cassandra client connection.
type CassandraClient struct {
	casendra *cassandra.CassandraConfig
	// Add real client fields here, e.g. session, cluster config etc.
}

// NewCassandraClient initializes a new Cassandra client.
// Replace with real initialization logic.
func NewCassandraClient(casendra *cassandra.CassandraConfig) *CassandraClient {
	return &CassandraClient{casendra: casendra}
}

// CassandraRepository implements the ICassandraRepository interface.
type CassandraRepository struct {
	client *CassandraClient
}

// NewCassandraRepository creates a new CassandraRepository with the given client.
func NewCassandraRepository(client *CassandraClient) repository.ICassandraRepository {
	return &CassandraRepository{client: client}
}

// GetData fetches data by key from Cassandra.

func (r *CassandraRepository) GetDataByID(idStr string) (*apiDtos.GenerateUrlRequest, error) {
	// Validate UUID
	id, err := authUtils.ParseUUID(idStr)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}
	// id is now a gocql.UUID

	// Check Cassandra session
	if r.client == nil || r.client.casendra == nil || r.client.casendra.Session == nil {
		return nil, errors.New("cassandra session is not properly initialized")
	}

	fmt.Printf("CassandraRepository: Getting data for ID %s\n", id.String())

	// Prepare variables to hold the result
	var (
		userID, name, requestType, modelType, email, phone, channelType string
		expireIn                                                        string
		otpRequired                                                     bool
		dataStr                                                         string
	)

	// Execute the query
	err = r.client.casendra.Session.Query(
		`SELECT user_id, name, request_type, model_type, email,
			expire_in, otp_required, phone, channel_type, data
		FROM protectedLink WHERE id = ? LIMIT 1`, id).Consistency(gocql.One).Scan(
		&userID, &name, &requestType, &modelType, &email,
		&expireIn, &otpRequired, &phone, &channelType, &dataStr,
	)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	// Deserialize JSON data field
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize data field: %w", err)
	}

	// Construct and return the DTO
	dto := &apiDtos.GenerateUrlRequest{
		UserID:      userID,
		Name:        name,
		RequestType: requestType,
		ModelType:   modelType,
		Email:       email,
		ExpireIn:    expireIn,
		OtpRequired: otpRequired,
		Phone:       phone,
		ChannelType: channelType,
		Data:        data,
	}

	return dto, nil
}

// SaveData saves data by key into Cassandra.
// SaveData saves the given GenerateUrlRequest DTO into Cassandra.
func (r *CassandraRepository) SaveData(dto *apiDtos.GenerateUrlRequest) (gocql.UUID, error) {
	// Validate required fields
	if dto.UserID == "" || dto.Name == "" || dto.RequestType == "" || dto.Email == "" {
		return gocql.UUID{}, errors.New("missing required fields")
	}

	// Check Cassandra client and session
	if r.client == nil || r.client.casendra == nil || r.client.casendra.Session == nil {
		return gocql.UUID{}, errors.New("Cassandra session is not properly initialized")
	}

	// Generate a new UUID for the record
	id := gocql.TimeUUID()

	// Serialize the `data` field to JSON
	dataBytes, err := json.Marshal(dto.Data)
	if err != nil {
		return gocql.UUID{}, fmt.Errorf("failed to serialize data: %w", err)
	}

	fmt.Printf("CassandraRepository: Saving data for user_id %s with id %s\n", dto.UserID, id.String())

	// Insert the record into Cassandra
	err = r.client.casendra.Session.Query(
		`INSERT INTO protectedLink (
			id, user_id, name, request_type, model_type, email,
			expire_in, otp_required, phone, channel_type, data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, dto.UserID, dto.Name, dto.RequestType, dto.ModelType, dto.Email,
		dto.ExpireIn, dto.OtpRequired, dto.Phone, dto.ChannelType, string(dataBytes),
	).Exec()

	if err != nil {
		return gocql.UUID{}, fmt.Errorf("failed to save data to Cassandra: %w", err)
	}

	fmt.Printf("CassandraRepository: Successfully saved data for user_id %s with id %s\n", dto.UserID, id.String())
	return id, nil
}

func (r *CassandraRepository) DeleteById(id string) error {
	// Check Cassandra client and session
	if r.client == nil || r.client.casendra == nil || r.client.casendra.Session == nil {
		return errors.New("cassandra session is not properly initialized")
	}

	// Delete the record from Cassandra
	err := r.client.casendra.Session.Query(
		`DELETE FROM protectedLink WHERE id = ?`,
		id,
	).Exec()

	if err != nil {
		return fmt.Errorf("failed to delete data from Cassandra: %w", err)
	}

	fmt.Printf("CassandraRepository: Successfully deleted data for id %s\n", id)
	return nil
}
