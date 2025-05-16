package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/models"
	authRepository "protected_link/internal/modules/authentication/repositories"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"

	"github.com/gocql/gocql"
)

// -------------------- Mock OTPRepository --------------------

type MockOTPRepository struct {
	mock.Mock
}

func (m *MockOTPRepository) SendOtp(request apiDtos.GenerateUrlRequest, otp, dbId string) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(request, otp, dbId)
	return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
}

func (m *MockOTPRepository) VerifyOtp(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(request)
	return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
}

func (m *MockOTPRepository) GetOTP(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

// -------------------- Mock CassandraRepository --------------------

type MockCassandraRepository struct {
	mock.Mock
}

func (m *MockCassandraRepository) GetDataByID(id string) (*apiDtos.GenerateUrlRequest, error) {
	args := m.Called(id)
	return args.Get(0).(*apiDtos.GenerateUrlRequest), args.Error(1)
}

func (m *MockCassandraRepository) DeleteById(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCassandraRepository) SaveData(dto *apiDtos.GenerateUrlRequest) (gocql.UUID, error) {
	args := m.Called(dto)
	return args.Get(0).(gocql.UUID), args.Error(1)
}

// -------------------- Test SendOtp --------------------

func TestSendOtp_Success(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	request := &apiDtos.GenerateUrlRequest{
		UserID:    "user123",
		Email:     "test@example.com",
		ModelType: "standard",
	}

	dbId := "db123"
	expectedResp := &commonDtos.ApiResponseDto{
		Message: "OTP sent successfully",
		Success: true,
	}

	mockOTPRepo.
		On("SendOtp", *request, mock.AnythingOfType("string"), dbId).
		Return(expectedResp, nil)

	resp, err := service.SendOtp(request, dbId)

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "OTP sent successfully", resp.Message)
	mockOTPRepo.AssertExpectations(t)
}

// -------------------- Test SendOtp Nil Request --------------------

func TestSendOtp_NilRequest(t *testing.T) {
	service := NewAuthenticationService(&authRepository.OTPRepository{}, &MockCassandraRepository{})

	resp, err := service.SendOtp(nil, "db123")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "request cannot be nil", err.Error())
}

// -------------------- Test SendOtp Error --------------------

func TestSendOtp_Error(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	request := &apiDtos.GenerateUrlRequest{
		UserID:    "user123",
		Email:     "test@example.com",
		ModelType: "standard",
	}

	dbId := "db123"
	expectedError := errors.New("failed to send OTP")

	mockOTPRepo.
		On("SendOtp", *request, mock.AnythingOfType("string"), dbId).
		Return((*commonDtos.ApiResponseDto)(nil), expectedError)

	resp, err := service.SendOtp(request, dbId)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, expectedError)
	mockOTPRepo.AssertExpectations(t)
}

// -------------------- Test VerifyOTP Success --------------------

func TestVerifyOtp_Success(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	request := &models.VerifyOTPRequest{
		VerificationID: "otp123",
		UserID:         "user123",
		OTP:            "456789",
	}

	expectedResp := &commonDtos.ApiResponseDto{
		Message: "OTP verified successfully",
		Success: true,
	}

	mockOTPRepo.On("VerifyOtp", request).Return(expectedResp, nil)

	resp, err := service.VerifyOTP(request)

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "OTP verified successfully", resp.Message)
	mockOTPRepo.AssertExpectations(t)
}

// -------------------- Test VerifyOTP Error --------------------

func TestVerifyOtp_Error(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	request := &models.VerifyOTPRequest{
		VerificationID: "otp123",
		UserID:         "user123",
		OTP:            "456789",
	}

	expectedError := errors.New("invalid OTP")

	mockOTPRepo.On("VerifyOtp", request).Return((*commonDtos.ApiResponseDto)(nil), expectedError)

	resp, err := service.VerifyOTP(request)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, expectedError)
	mockOTPRepo.AssertExpectations(t)
}

// -------------------- Test GetAuthToken Success --------------------

func TestGetAuthToken_Success(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	expectedData := &apiDtos.GenerateUrlRequest{
		UserID:    "user123",
		ModelType: "standard",
	}

	mockCassandra.On("GetDataByID", "token123").Return(expectedData, nil)

	result, err := service.GetAuthToken("user123", "token123")

	assert.NoError(t, err)
	assert.Equal(t, "user123", result.UserID)
	mockCassandra.AssertExpectations(t)
}

// -------------------- Test GetAuthToken Error --------------------

func TestGetAuthToken_NilRepository(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	service := NewAuthenticationService(mockOTPRepo, nil)

	result, err := service.GetAuthToken("user123", "token123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "cassandra repository not initialized", err.Error())
}

func TestGetAuthToken_RepositoryError(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	expectedError := errors.New("database error")
	mockCassandra.On("GetDataByID", "token123").Return((*apiDtos.GenerateUrlRequest)(nil), expectedError)

	result, err := service.GetAuthToken("user123", "token123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get auth token")
	mockCassandra.AssertExpectations(t)
}

// -------------------- Test VerifyOTP Nil Request --------------------

func TestVerifyOtp_NilRequest(t *testing.T) {
	mockOTPRepo := new(MockOTPRepository)
	mockCassandra := new(MockCassandraRepository)

	service := NewAuthenticationService(mockOTPRepo, mockCassandra)

	resp, err := service.VerifyOTP(nil)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "request cannot be nil", err.Error())
}

// -------------------- Test VerifyOTP Nil Repository --------------------

func TestVerifyOtp_NilRepository(t *testing.T) {
	service := NewAuthenticationService(nil, &MockCassandraRepository{})

	request := &models.VerifyOTPRequest{
		VerificationID: "otp123",
		UserID:         "user123",
		OTP:            "456789",
	}

	resp, err := service.VerifyOTP(request)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "OTP repository not initialized", err.Error())
}

// -------------------- Test SendOtp Nil Repository --------------------

func TestSendOtp_NilRepository(t *testing.T) {
	service := NewAuthenticationService(nil, &MockCassandraRepository{})

	request := &apiDtos.GenerateUrlRequest{
		UserID:    "user123",
		Email:     "test@example.com",
		ModelType: "standard",
	}

	resp, err := service.SendOtp(request, "db123")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "OTP repository not initialized", err.Error())
}
