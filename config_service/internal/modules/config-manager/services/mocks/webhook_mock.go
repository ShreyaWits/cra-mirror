package mocks // It's common to put mocks in a dedicated 'mocks' package

import (
	"github.com/stretchr/testify/mock" // Import the mock library

	// Import the DTOs and the interface you are mocking
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	// Assuming IWebhookService is in this package
)

// MockWebhookService is a mock implementation of the services.IWebhookService interface.
// It embeds mock.Mock to provide the core mocking capabilities.
type MockWebhookService struct {
	mock.Mock
}

// RegisterWebhookService is a mock implementation of the corresponding IWebhookService method.
// It records the call and returns the values set in the test.
func (m *MockWebhookService) RegisterWebhookService(req dtos.RegisterWebhookRequest) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	// Record the call with the provided arguments.
	args := m.Called(req)

	// Retrieve the return values that were set using .Return() in the test.
	// Ensure correct type casting for each expected return value.
	var successResp dtos.SuccessResponse
	if args.Get(0) != nil {
		successResp = args.Get(0).(dtos.SuccessResponse)
	}

	var serviceErrResp *dtos.ServiceErrorResponse
	if args.Get(1) != nil {
		serviceErrResp = args.Get(1).(*dtos.ServiceErrorResponse)
	}

	// Return the retrieved values.
	return &successResp, serviceErrResp
}

// GetWebhooks is a mock implementation of the corresponding IWebhookService method.
func (m *MockWebhookService) GetWebhooks(env, service string) ([]dtos.RegisterWebhookRequest, *dtos.ServiceErrorResponse) {
	// Record the call with the provided arguments.
	args := m.Called(env, service)

	// Retrieve the return values.
	var webhooks []dtos.RegisterWebhookRequest
	if args.Get(0) != nil {
		// Cast the first return value to the expected slice type.
		webhooks = args.Get(0).([]dtos.RegisterWebhookRequest)
	}

	var serviceErrResp *dtos.ServiceErrorResponse
	if args.Get(1) != nil {
		serviceErrResp = args.Get(1).(*dtos.ServiceErrorResponse)
	}

	return webhooks, serviceErrResp
}

// DeleteWebhook is a mock implementation of the corresponding IWebhookService method.
func (m *MockWebhookService) DeleteWebhook(env, service, url, method string) (string, *dtos.ServiceErrorResponse) {
	// Record the call.
	args := m.Called(env, service, url, method)

	// Retrieve the return values.
	var message string
	if args.Get(0) != nil {
		message = args.Get(0).(string)
	}

	var serviceErrResp *dtos.ServiceErrorResponse
	if args.Get(1) != nil {
		serviceErrResp = args.Get(1).(*dtos.ServiceErrorResponse)
	}

	return message, serviceErrResp
}

// DeleteAllWebhooks is a mock implementation of the corresponding IWebhookService method.
func (m *MockWebhookService) DeleteAllWebhooks(env, service string) error {
	// Record the call.
	args := m.Called(env, service)

	// Retrieve the error return value.
	return args.Error(0)
}

// NotifyWebhook is a mock implementation of the corresponding IWebhookService method.
// It has no return values, so we just record the call.
func (m *MockWebhookService) NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
	// Record the call.
	m.Called(hook, data)
	// No return values, so nothing to retrieve.
}

// --- Example Usage in a Test ---
/*
import (
	"testing"
	"github.com/stretchr/testify/assert"

	// Import your service and the mock package
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/internal/modules/config-manager/services/mocks" // Assuming your mocks are here

	// Import DTOs needed for test setup
	"nps-config-service/internal/modules/config-manager/apis/dtos"
)

func TestYourServiceMethod(t *testing.T) {
	// Create an instance of the mock webhook service
	mockWebhookService := new(mocks.MockWebhookService)

	// Create an instance of the service you are testing, injecting the mock
	// Assuming your service struct and constructor look like this:
	// type YourService struct { WebhookService services.IWebhookService }
	// func NewYourService(webhookService services.IWebhookService) *YourService { ... }
	// yourServiceInstance := services.NewYourService(mockWebhookService)

	// Define the expected behavior of the mock
	// Example: Expect GetWebhooks to be called once with specific arguments
	expectedWebhooks := []dtos.RegisterWebhookRequest{
		{URL: "http://example.com/hook1", Method: "POST", Environment: "dev", ServiceName: "myservice"},
	}
	mockWebhookService.On("GetWebhooks", "dev", "myservice").Return(expectedWebhooks, nil).Once()

	// Example: Expect NotifyWebhook to be called for each hook found
	// mockWebhookService.On("NotifyWebhook", mock.AnythingOfType("dtos.RegisterWebhookRequest"), mock.AnythingOfType("map[string]interface{}")).Return().Times(len(expectedWebhooks))


	// Call the method on your service that uses the webhook service
	// For example:
	// err := yourServiceInstance.ProcessSomethingThatUsesWebhooks("dev", "myservice", someData)

	// Assert the results of your service method call
	// assert.NoError(t, err, "Service method should not return an error")

	// Assert that the expected methods on the mock were called
	mockWebhookService.AssertExpectations(t)
}
*/
