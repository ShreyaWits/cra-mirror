package mocks

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/stretchr/testify/mock"
)

type MockConfigService struct {
	mock.Mock
}

func (m *MockConfigService) StoreConfigService(env string, service string, req map[string]interface{}) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	args := m.Called(env, service, req)
	if args.Error(1) != nil {
		res, _:= args.Get(1).(*dtos.ServiceErrorResponse)

		return nil, res
	}
	res, _ := args.Get(0).(*dtos.SuccessResponse) // Type assertion
	repposeErr, _:= args.Get(1).(*dtos.ServiceErrorResponse)
	return res, repposeErr
	// return args.Get(0), nil
}

func (m *MockConfigService) GetConfigService(service string, env string) (interface{}, error) {
	args := m.Called(service, env)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0), nil
}

func (m *MockConfigService) GetConfigValueService(serviceName string, env string, key string) (interface{}, error) {
	args := m.Called(serviceName, env, key)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0), nil
}

func (m *MockConfigService) GetConfigMetadataService(serviceName string, env string) (interface{}, error) {
	args := m.Called(serviceName, env)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0), nil
}
