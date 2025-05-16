package infrastructure

import (
	"encoding/json"
	"errors"
	"testing"

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

func TestSaveData_Success(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ExecFunc: func() error { return nil },
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)
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
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)
	dto, err := repo.GetDataByID("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")
	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, "user123", dto.UserID)
	assert.Equal(t, "test", dto.Name)
	assert.Equal(t, expectedData["key"], dto.Data["key"])
}

func TestDeleteById_Success(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ExecFunc: func() error { return nil },
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)
	err := repo.DeleteById("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")
	assert.NoError(t, err)
}

func TestSaveData_MissingFields(t *testing.T) {
	mockConfig := &cassandra.CassandraConfig{}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)
	dto := &apiDtos.GenerateUrlRequest{
		Name: "test", // missing UserID, Email, RequestType
	}
	_, err := repo.SaveData(dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")
}

func TestGetDataByID_InvalidUUID(t *testing.T) {
	mockConfig := &cassandra.CassandraConfig{}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)

	dto, err := repo.GetDataByID("invalid-uuid")
	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Equal(t, "invalid UUID format", err.Error())
}

func TestGetDataByID_SessionNil(t *testing.T) {
	repo := NewCassandraRepository(&CassandraClient{casendra: nil})
	dto, err := repo.GetDataByID("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")
	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Equal(t, "cassandra session is not properly initialized", err.Error())
}

func TestGetDataByID_ScanError(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ScanFunc: func(dest ...interface{}) error {
					return gocql.ErrUnavailable
				},
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)

	dto, err := repo.GetDataByID("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")
	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "failed to fetch data")
}

func TestGetDataByID_UnmarshalError(t *testing.T) {
	badJSON := "{not-valid-json"

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
					*dest[9].(*string) = badJSON
					return nil
				},
			}
		},
	}

	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)

	dto, err := repo.GetDataByID("b3cb2e24-6f0e-11ed-a1eb-0242ac120002")
	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "failed to deserialize data field")
}

func TestSaveData_SessionNil(t *testing.T) {
	repo := NewCassandraRepository(&CassandraClient{casendra: nil})
	dto := &apiDtos.GenerateUrlRequest{
		UserID:      "user123",
		Name:        "test",
		RequestType: "req",
		Email:       "test@example.com",
		Data:        map[string]interface{}{"key": "value"},
	}
	_, err := repo.SaveData(dto)
	assert.Error(t, err)
	assert.Equal(t, "Cassandra session is not properly initialized", err.Error())
}

func TestSaveData_JSONMarshalError(t *testing.T) {
	repo := NewCassandraRepository(&CassandraClient{casendra: &cassandra.CassandraConfig{
		Session: &MockSession{
			QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
				return &MockQuery{
					ExecFunc: func() error { return nil },
				}
			},
		},
	}})
	dto := &apiDtos.GenerateUrlRequest{
		UserID:      "user123",
		Name:        "test",
		RequestType: "req",
		Email:       "test@example.com",
		Data:        map[string]interface{}{"invalid": make(chan int)}, // not serializable
	}
	_, err := repo.SaveData(dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to serialize data")
}

func TestSaveData_QueryExecError(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ExecFunc: func() error { return errors.New("insert failed") },
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)

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
	_, err := repo.SaveData(dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save data to Cassandra")
}

func TestDeleteById_SessionNil(t *testing.T) {
	repo := NewCassandraRepository(&CassandraClient{casendra: nil})
	err := repo.DeleteById("some-id")
	assert.Error(t, err)
	assert.Equal(t, "cassandra session is not properly initialized", err.Error())
}

func TestDeleteById_ExecError(t *testing.T) {
	mockSession := &MockSession{
		QueryFunc: func(string, ...interface{}) cassandra.QueryInterface {
			return &MockQuery{
				ExecFunc: func() error { return errors.New("delete failed") },
			}
		},
	}
	mockConfig := &cassandra.CassandraConfig{Session: mockSession}
	client := NewCassandraClient(mockConfig)
	repo := NewCassandraRepository(client)

	err := repo.DeleteById("some-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete data from Cassandra")
}
