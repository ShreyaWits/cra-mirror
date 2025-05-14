package handler

// import (
// 	"context"
// 	"template-services/internal/models"
// 	"testing"

// 	"github.com/google/uuid"
// )

// // Compile-time assertion to ensure a type implements TemplateServiceInterface
// // var _ TemplateServiceInterface = (*MockTemplateService)(nil)

// // MockTemplateService is a mock implementation of TemplateServiceInterface for testing.
// type MockTemplateService struct {
// 	CreateTemplateFn  func(ctx context.Context, template *models.Template) (*models.Template, error)
// 	GetTemplateFn     func(ctx context.Context, id, name, channel, language string) (*models.Template, error)
// 	GetTemplateByIDFn func(ctx context.Context, id uuid.UUID) (*models.Template, error)
// 	UpdateTemplateFn  func(ctx context.Context, template *models.Template) (*models.Template, error)
// 	DeleteTemplateFn  func(ctx context.Context, id string) (*models.Template, error)
// }

// func (m *MockTemplateService) On(s string, ctx context.Context, param3 string) {
// 	panic("unimplemented")
// }

// func (m *MockTemplateService) AssertExpectations(t *testing.T) {
// 	panic("unimplemented")
// }

// func (m *MockTemplateService) CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
// 	return m.CreateTemplateFn(ctx, template)
// }
// func (m *MockTemplateService) GetTemplate(ctx context.Context, id, name, channel, language string) (*models.Template, error) {
// 	return m.GetTemplateFn(ctx, id, name, channel, language)
// }
// func (m *MockTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
// 	return m.GetTemplateByIDFn(ctx, id)
// }
// func (m *MockTemplateService) UpdateTemplate(ctx context.Context, template *models.Template) (*models.Template, error) {
// 	return m.UpdateTemplateFn(ctx, template)
// }
// func (m *MockTemplateService) DeleteTemplate(ctx context.Context, id string) (*models.Template, error) {
// 	return m.DeleteTemplateFn(ctx, id)
// }

// package handler

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"

// 	"template-services/internal/models"
// 	"template-services/internal/template/dto"
// 	pb "template-services/proto"
// )

// // MockTemplateService mocks the template service interface used by the handler.
// type MockTemplateService struct {
// 	mock.Mock
// }

// func (m *MockTemplateService) GetTemplate(ctx context.Context, templateID string) (*dto.TemplateResponse, error) {
// 	args := m.Called(ctx, templateID)
// 	// Defensive: handle nil return for *dto.TemplateResponse
// 	var resp *dto.TemplateResponse
// 	if tmp := args.Get(0); tmp != nil {
// 		resp = tmp.(*dto.TemplateResponse)
// 	}
// 	return resp, args.Error(1)
// }

// // Add stub/mock implementations for all other methods required by the interface.
// func (m *MockTemplateService) CreateTemplate(ctx context.Context, req *models.Template) (*models.Template, error) {
// 	args := m.Called(ctx, req)
// 	var resp *models.Template
// 	if tmp := args.Get(0); tmp != nil {
// 		resp = tmp.(*models.Template)
// 	}
// 	return resp, args.Error(1)
// }

// func (m *MockTemplateService) UpdateTemplate(ctx context.Context, req *dto.UpdateTemplateRequest) (*dto.TemplateResponse, error) {
// 	args := m.Called(ctx, req)
// 	var resp *dto.TemplateResponse
// 	if tmp := args.Get(0); tmp != nil {
// 		resp = tmp.(*dto.TemplateResponse)
// 	}
// 	return resp, args.Error(1)
// }

// func (m *MockTemplateService) DeleteTemplate(ctx context.Context, templateID string) (*models.Template, error) {
// 	args := m.Called(ctx, templateID)
// 	var resp *models.Template
// 	if tmp := args.Get(0); tmp != nil {
// 		resp = tmp.(*models.Template)
// 	}
// 	return resp, args.Error(1)
// }

// func (m *MockTemplateService) ListTemplates(ctx context.Context, filter *dto.ListTemplatesFilter) ([]*dto.TemplateResponse, error) {
// 	args := m.Called(ctx, filter)
// 	var resp []*dto.TemplateResponse
// 	if tmp := args.Get(0); tmp != nil {
// 		resp = tmp.([]*dto.TemplateResponse)
// 	}
// 	return resp, args.Error(1)
// }

// func newTestHandler() (*TemplateGRPCHandler, *MockTemplateService) {
// 	mockSvc := new(MockTemplateService)
// 	handler := &TemplateGRPCHandler{
// 		service: mockSvc,
// 	}
// 	return handler, mockSvc
// }

// func TestTemplateGRPCHandler_GetTemplate_Success(t *testing.T) {
// 	handler, mockSvc := newTestHandler()
// 	ctx := context.Background()
// 	req := &pb.GetTemplateRequest{TemplateId: "template-1"}
// 	dtoResp := &dto.TemplateResponse{ID: "template-1", Name: "Test Template"}

// 	mockSvc.On("GetTemplate", ctx, "template-1").Return(dtoResp, nil)

// 	resp, err := handler.GetTemplate(ctx, req)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.Equal(t, "template-1", resp.ID)
// 	assert.Equal(t, "Test Template", resp.Name)
// 	mockSvc.AssertExpectations(t)
// }

// func TestTemplateGRPCHandler_GetTemplate_NotFound(t *testing.T) {
// 	handler, mockSvc := newTestHandler()
// 	ctx := context.Background()
// 	req := &pb.GetTemplateRequest{TemplateId: "template-1"}

// 	mockSvc.On("GetTemplate", ctx, "template-1").Return((*dto.TemplateResponse)(nil), errors.New("not found"))

// 	resp, err := handler.GetTemplate(ctx, req)
// 	assert.Nil(t, resp)
// 	assert.Error(t, err)
// 	st, ok := status.FromError(err)
// 	assert.True(t, ok)
// 	assert.Equal(t, codes.NotFound, st.Code())
// 	mockSvc.AssertExpectations(t)
// }

// func TestTemplateGRPCHandler_GetTemplate_NilRequest(t *testing.T) {
// 	handler, _ := newTestHandler()
// 	ctx := context.Background()

// 	resp, err := handler.GetTemplate(ctx, nil)
// 	assert.Nil(t, resp)
// 	assert.Error(t, err)
// 	st, ok := status.FromError(err)
// 	assert.True(t, ok)
// 	assert.Equal(t, codes.InvalidArgument, st.Code())
// }
