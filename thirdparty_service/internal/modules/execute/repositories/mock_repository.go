package repositories

// MockRepository is a mock implementation of the Repository interface for testing
type MockRepository struct{}

func (m *MockRepository) VerifyAadhaar(aadhaar string) (bool, string, string, error) {
	return true, "John Doe", "1990-01-01", nil
}

func (m *MockRepository) VerifyPAN(pan string) (bool, string, string, error) {
	return true, "John Doe", "Individual", nil
}

func (m *MockRepository) SendSMS(phone, message string) (string, error) {
	return "sent", nil
}

func (m *MockRepository) SendEmail(to, subject, body string) (string, error) {
	return "sent", nil
}

func (m *MockRepository) InitiatePayment(userID string, amount float64) (string, string, error) {
	return "txn_" + userID, "success", nil
}
