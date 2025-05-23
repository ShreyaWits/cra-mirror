package repositories

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRepo_VerifyAadhaar(t *testing.T) {
	mockRepo := new(MockRepo)
	aadhaar := "123412341234"
	expectedValid := true
	expectedName := "John Doe"
	expectedDOB := "1990-01-01"
	expectedErr := errors.New("some error")

	// Test case 1: Successful verification
	mockRepo.On("VerifyAadhaar", aadhaar).Return(expectedValid, expectedName, expectedDOB, nil).Once()
	valid, name, dob, err := mockRepo.VerifyAadhaar(aadhaar)
	assert.Equal(t, expectedValid, valid)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedDOB, dob)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Reset mock for next test case
	mockRepo = new(MockRepo)

	// Test case 2: Verification with error
	mockRepo.On("VerifyAadhaar", aadhaar).Return(false, "", "", expectedErr).Once()
	valid, name, dob, err = mockRepo.VerifyAadhaar(aadhaar)
	assert.False(t, valid)
	assert.Equal(t, "", name)
	assert.Equal(t, "", dob)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestMockRepo_VerifyPAN(t *testing.T) {
	mockRepo := new(MockRepo)
	pan := "ABCDE1234F"
	expectedValid := true
	expectedName := "John Doe"
	expectedErr := errors.New("some error")

	// Test case 1: Successful verification
	mockRepo.On("VerifyPAN", pan).Return(expectedValid, expectedName, "", nil).Once()
	valid, name,_, err := mockRepo.VerifyPAN(pan)
	assert.Equal(t, expectedValid, valid)
	assert.Equal(t, expectedName, name)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Reset mock for next test case
	mockRepo = new(MockRepo)

	// Test case 2: Verification with error
	mockRepo.On("VerifyPAN", pan).Return(false, "", "", expectedErr).Once()
	valid, name,_, err = mockRepo.VerifyPAN(pan)
	assert.False(t, valid)
	assert.Equal(t, "", name)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestMockRepo_SendSMS(t *testing.T) {
	mockRepo := new(MockRepo)
	phone := "1234567890"
	message := "Test message"
	expectedStatus := "sent"
	expectedErr := errors.New("some error")

	// Test case 1: Successful send
	mockRepo.On("SendSMS", phone, message).Return(expectedStatus, nil).Once()
	status, err := mockRepo.SendSMS(phone, message)
	assert.Equal(t, expectedStatus, status)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Reset mock for next test case
	mockRepo = new(MockRepo)

	// Test case 2: Send with error
	mockRepo.On("SendSMS", phone, message).Return("", expectedErr).Once()
	status, err = mockRepo.SendSMS(phone, message)
	assert.Equal(t, "", status)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestMockRepo_SendEmail(t *testing.T) {
	mockRepo := new(MockRepo)
	to := "test@example.com"
	subject := "Test Subject"
	body := "Test Body"
	expectedStatus := "sent"
	expectedErr := errors.New("some error")

	// Test case 1: Successful send
	mockRepo.On("SendEmail", to, subject, body).Return(expectedStatus, nil).Once()
	status, err := mockRepo.SendEmail(to, subject, body)
	assert.Equal(t, expectedStatus, status)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Reset mock for next test case
	mockRepo = new(MockRepo)

	// Test case 2: Send with error
	mockRepo.On("SendEmail", to, subject, body).Return("", expectedErr).Once()
	status, err = mockRepo.SendEmail(to, subject, body)
	assert.Equal(t, "", status)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestMockRepo_InitiatePayment(t *testing.T) {
	mockRepo := new(MockRepo)
	userID := "user123"
	amount := 100.50
	expectedTransactionID := "txn_user123"
	expectedStatus := "success"
	expectedErr := errors.New("some error")

	// Test case 1: Successful initiation
	mockRepo.On("InitiatePayment", userID, amount).Return(expectedTransactionID, expectedStatus, nil).Once()
	transactionID, status, err := mockRepo.InitiatePayment(userID, amount)
	assert.Equal(t, expectedTransactionID, transactionID)
	assert.Equal(t, expectedStatus, status)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Reset mock for next test case
	mockRepo = new(MockRepo)

	// Test case 2: Initiation with error
	mockRepo.On("InitiatePayment", userID, amount).Return("", "", expectedErr).Once()
	transactionID, status, err = mockRepo.InitiatePayment(userID, amount)
	assert.Equal(t, "", transactionID)
	assert.Equal(t, "", status)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}
