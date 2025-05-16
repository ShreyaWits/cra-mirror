package repository_test

import (
	"encoding/json"
	"testing"

	"protected_link/internal/modules/cassandra/infrastructure"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/pkg/cassandra"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

// --- MockQuery for cassandra.QueryInterface ---
type MockQuery struct {
	ScanFunc func(...interface{}) error
	ExecFunc func() error
}

func (m *MockQuery) Consistency(_ gocql.Consistency) cassandra.QueryInterface {
	return m
}

func (m *MockQuery) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

func (m *MockQuery) Exec() error {
	if m.ExecFunc != nil {
		return m.ExecFunc()
	}
	return nil
}

// --- MockSession for cassandra.SessionInterface ---
type MockSession struct {
	QueryFunc func(string, ...interface{}) cassandra.QueryInterface
	CloseFunc func()
}

func (m *MockSession) Query(stmt string, args ...interface{}) cassandra.QueryInterface {
	return m.QueryFunc(stmt, args...)
}

func (m *MockSession) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// Test: Save Data (Success)
func TestSaveData_Success(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ExecFunc: func() error { return nil },
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := infrastructure.NewCassandraClient(mockConfig)
	repo := infrastructure.NewCassandraRepository(client)

	dto := &apiDtos.GenerateUrlRequest{
		UserID:      "user123",
		Name:        "test name",
		RequestType: "type1",
		ModelType:   "modelX",
		Email:       "test@example.com",
		Phone:       "1234567890",
		ExpireIn:    "30m",
		OtpRequired: true,
		ChannelType: "SMS",
		Data:        map[string]interface{}{"key": "value"},
	}

	id, err := repo.SaveData(dto)

	assert.NoError(t, err)
	assert.NotEqual(t, gocql.UUID{}, id)
}

// Test: Get Data by ID (Success)
func TestGetDataByID_Success(t *testing.T) {
	expectedData := map[string]interface{}{"key": "value"}
	dataStr, _ := json.Marshal(expectedData)

	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ScanFunc: func(dest ...interface{}) error {
					*dest[0].(*string) = "user123"
					*dest[1].(*string) = "test"
					*dest[2].(*string) = "type1"
					*dest[3].(*string) = "model"
					*dest[4].(*string) = "email@example.com"
					*dest[5].(*string) = "30m"
					*dest[6].(*bool) = true
					*dest[7].(*string) = "123456"
					*dest[8].(*string) = "channel"
					*dest[9].(*string) = string(dataStr)
					return nil
				},
			}
		},
	}

	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := infrastructure.NewCassandraClient(mockConfig)
	repo := infrastructure.NewCassandraRepository(client)

	dto, err := repo.GetDataByID("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, "user123", dto.UserID)
	assert.Equal(t, "test", dto.Name)
	assert.Equal(t, expectedData["key"], dto.Data["key"])
}

// Test: Delete Data by ID (Success)
func TestDeleteById_Success(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ExecFunc: func() error {
					return nil
				},
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := infrastructure.NewCassandraClient(mockConfig)
	repo := infrastructure.NewCassandraRepository(client)

	err := repo.DeleteById("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")
	assert.NoError(t, err)
}

// Test: Save Data with Missing Fields
func TestSaveData_MissingFields(t *testing.T) {
	mockConfig := &cassandra.CassandraConfig{}
	client := infrastructure.NewCassandraClient(mockConfig)
	repo := infrastructure.NewCassandraRepository(client)

	dto := &apiDtos.GenerateUrlRequest{
		Name: "test", // missing UserID, Email, RequestType
	}

	_, err := repo.SaveData(dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")
}
