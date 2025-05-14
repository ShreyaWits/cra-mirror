package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"template-services/internal/models"
	appErrors "template-services/internal/pkg/errors"
	pb "template-services/proto"
)

// Mock for TemplateServiceInterface

type MockTemplateService struct {
	mock.Mock
}

// CreateTemplate implements service.TemplateServiceInterface.
func (m *MockTemplateService) CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	panic("unimplemented")
}

// DeleteTemplate implements service.TemplateServiceInterface.
func (m *MockTemplateService) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
	panic("unimplemented")
}

// GetTemplateByID implements service.TemplateServiceInterface.
func (m *MockTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	panic("unimplemented")
}

// UpdateTemplate implements service.TemplateServiceInterface.
func (m *MockTemplateService) UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
	panic("unimplemented")
}




func (m *MockTemplateService) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
	args := m.Called(ctx, id, name, channel, language)
	resp := args.Get(0)
	if resp == nil {
		return nil, args.Error(1)
	}
	return resp.(*models.Template), args.Error(1)
}

func TestGetTemplateV1_Success(t *testing.T) {
	mockSvc := new(MockTemplateService)
	handler := NewTemplateGRPCHandler(mockSvc)

	req := &pb.GetTemplateRequest{
		Name:     "TestTemplate",
		Channel:  "email",
		Language: "en",
	}

	expectedResp := &models.Template{
		ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Name:     "TestTemplate",
		Channel:  "email",
		Language: "en",
	}

	mockSvc.On("GetTemplate", mock.Anything, "", "TestTemplate", "email", "en").Return(expectedResp, nil)

	resp, err := handler.GetTemplateV1(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Template retrieved successfully", resp.Message["info"])
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", resp.Data["id"])
	mockSvc.AssertExpectations(t)
}

func TestGetTemplateV1_InvalidRequest(t *testing.T) {
	mockSvc := new(MockTemplateService)
	handler := NewTemplateGRPCHandler(mockSvc)

	req := &pb.GetTemplateRequest{
		Name:     "",
		Channel:  "email",
		Language: "en",
	}

	resp, err := handler.GetTemplateV1(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Message)
	assert.Equal(t, appErrors.TmpErrInvalidRequestBody, resp.Error["code"])
	mockSvc.AssertNotCalled(t, "GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetTemplateV1_TemplateNotFound(t *testing.T) {
	mockSvc := new(MockTemplateService)
	handler := NewTemplateGRPCHandler(mockSvc)

	req := &pb.GetTemplateRequest{
		Name:     "NotExist",
		Channel:  "email",
		Language: "en",
	}

	mockSvc.On("GetTemplate", mock.Anything, "", "NotExist", "email", "en").Return(nil, errors.New("not found"))

	resp, err := handler.GetTemplateV1(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, appErrors.TmpErrTemplateNotFound, resp.Error["code"])
	assert.Equal(t, appErrors.GetAppErrorMessage(appErrors.TmpErrTemplateNotFound), resp.Message["template"])
	mockSvc.AssertExpectations(t)
}
