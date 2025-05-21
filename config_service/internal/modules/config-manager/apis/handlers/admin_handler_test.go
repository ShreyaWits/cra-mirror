package handler

import (
	"encoding/json"
	"errors"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
)

type MockAdminService struct {
	mock.Mock
}

func (m *MockAdminService) CreateAdminService(admin *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error) {
	args := m.Called(admin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ResponseAdminSignupDto), args.Error(1)
}

func (m *MockAdminService) FetchAdminService(admin *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error) {
	args := m.Called(admin, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ResponseAdminDto), args.Error(1)
}

var (
	ErrInvalidSecret    = errors.New("invalid admin secret")
	ErrMissingJWTSecret = errors.New("missing JWT secret")
)

// Middleware to bind request body to context
func bindRequestBody(c *fiber.Ctx) error {
	var data interface{}
	if err := c.BodyParser(&data); err != nil {
		return err
	}
	c.Locals("contextData", data)
	return c.Next()
}

func TestCreateAdminHandler(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAdminService)

	tests := []struct {
		name          string
		input         *dtos.AdminSignupDto
		mockResponse  *dtos.ResponseAdminSignupDto
		mockError     error
		expectedCode  int
		expectedError bool
		envVars       map[string]string
	}{
		{
			name: "Successful admin creation",
			input: &dtos.AdminSignupDto{
				Username: "testadmin",
				Password: "testpass",
				Secret:   "validsecret",
			},
			mockResponse: &dtos.ResponseAdminSignupDto{
				Success: true,
				Message: "Admin created successfully",
				AdminId: "testadmin",
			},
			expectedCode: fiber.StatusOK,
			envVars: map[string]string{
				"ADMIN_SECRET": "validsecret",
			},
		},
		{
			name: "Invalid secret",
			input: &dtos.AdminSignupDto{
				Username: "testadmin",
				Password: "testpass",
				Secret:   "invalidsecret",
			},
			mockResponse:  nil,
			mockError:     ErrInvalidSecret,
			expectedCode:  fiber.StatusBadRequest,
			expectedError: true,
			envVars: map[string]string{
				"ADMIN_SECRET": "validsecret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create handler with mock service
			handler := NewAdminHandler(mockService)

			// Create request body
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Setup mock expectations
			if !tt.expectedError {
				mockService.On("CreateAdminService", mock.MatchedBy(func(req *dtos.AdminSignupDto) bool {
					return req.Username == tt.input.Username &&
						req.Password == tt.input.Password &&
						req.Secret == tt.input.Secret
				})).Return(tt.mockResponse, tt.mockError).Once()
			}

			// Create a new fasthttp request context
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod("POST")
			ctx.Request.SetRequestURI("/admin/signup")
			ctx.Request.SetBody(body)
			ctx.Request.Header.SetContentType("application/json")

			// Create a new Fiber context
			fiberCtx := app.AcquireCtx(ctx)
			defer app.ReleaseCtx(fiberCtx)

			// Parse and bind request body
			var reqData dtos.AdminSignupDto
			err = json.Unmarshal(body, &reqData)
			assert.NoError(t, err)
			fiberCtx.Locals("contextData", &reqData)

			// Call handler directly
			err = handler.CreateAdminHandler(fiberCtx)

			// Get response
			resp := fiberCtx.Response()

			// Assertions
			assert.Equal(t, tt.expectedCode, resp.StatusCode())

			if !tt.expectedError {
				var response dtos.ResponseAdminSignupDto
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Success, response.Success)
				assert.Equal(t, tt.mockResponse.Message, response.Message)
				assert.Equal(t, tt.mockResponse.AdminId, response.AdminId)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestFetchAdminHandler(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAdminService)

	tests := []struct {
		name          string
		input         *dtos.AdminLoginDto
		secret        string
		mockResponse  *dtos.ResponseAdminDto
		mockError     error
		expectedCode  int
		expectedError bool
		envVars       map[string]string
	}{
		{
			name: "Successful admin login",
			input: &dtos.AdminLoginDto{
				Username: "testadmin",
				Password: "testpass",
			},
			secret: "testsecret",
			mockResponse: &dtos.ResponseAdminDto{
				Success:      true,
				Message:      "Login successful",
				AdminId:      "testadmin",
				Token:        "validtoken",
				RefreshToken: "validrefreshtoken",
			},
			expectedCode: fiber.StatusOK,
			envVars: map[string]string{
				"JWT_SECRET": "testsecret",
			},
		},
		{
			name: "Missing JWT secret",
			input: &dtos.AdminLoginDto{
				Username: "testadmin",
				Password: "testpass",
			},
			secret:        "",
			mockResponse:  nil,
			mockError:     ErrMissingJWTSecret,
			expectedCode:  fiber.StatusBadRequest,
			expectedError: true,
			envVars:       map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create handler with mock service
			handler := NewAdminHandler(mockService)

			// Create request body
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Setup mock expectations
			if !tt.expectedError {
				mockService.On("FetchAdminService", mock.MatchedBy(func(req *dtos.AdminLoginDto) bool {
					return req.Username == tt.input.Username &&
						req.Password == tt.input.Password
				}), tt.secret).Return(tt.mockResponse, tt.mockError).Once()
			}

			// Create a new fasthttp request context
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod("POST")
			ctx.Request.SetRequestURI("/admin/login")
			ctx.Request.SetBody(body)
			ctx.Request.Header.SetContentType("application/json")

			// Create a new Fiber context
			fiberCtx := app.AcquireCtx(ctx)
			defer app.ReleaseCtx(fiberCtx)

			// Parse and bind request body
			var reqData dtos.AdminLoginDto
			err = json.Unmarshal(body, &reqData)
			assert.NoError(t, err)
			fiberCtx.Locals("contextData", &reqData)

			// Call handler directly
			err = handler.FetchAdminHandler(fiberCtx)

			// Get response
			resp := fiberCtx.Response()

			// Assertions
			assert.Equal(t, tt.expectedCode, resp.StatusCode())

			if !tt.expectedError {
				var response dtos.ResponseAdminDto
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Success, response.Success)
				assert.Equal(t, tt.mockResponse.Message, response.Message)
				assert.Equal(t, tt.mockResponse.AdminId, response.AdminId)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
				assert.Equal(t, tt.mockResponse.RefreshToken, response.RefreshToken)
			}

			mockService.AssertExpectations(t)
		})
	}
}
