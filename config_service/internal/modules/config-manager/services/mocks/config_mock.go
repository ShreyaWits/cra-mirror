package mocks

import (
	"context"
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/stretchr/testify/mock"
)

type MockConfigService struct {
	mock.Mock
}

func (m *MockConfigService) StoreConfigService(ctx context.Context, env string, service string, req map[string]interface{}) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	args := m.Called(ctx, env, service, req)

	if args.Get(1) != nil {
		errResp, _ := args.Get(1).(*dtos.ServiceErrorResponse)
		return nil, errResp
	}

	successResp, _ := args.Get(0).(*dtos.SuccessResponse)
	return successResp, nil
}

func (m *MockConfigService) GetConfigService(ctx context.Context, service string, env string) (interface{}, error) {
	args := m.Called(ctx, service, env)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0), nil
}

func (m *MockConfigService) GetConfigValueService(ctx context.Context, serviceName string, env string, key string) (interface{}, error) {
	args := m.Called(ctx, serviceName, env, key)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0), nil
}

func (m *MockConfigService) GetConfigMetadataService(ctx context.Context, serviceName string, env string) (interface{}, error) {
	args := m.Called(ctx, serviceName, env)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0), nil
}
