package mocks

import (
	"context"
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/stretchr/testify/mock"
)

type MockAdminService struct {
	mock.Mock
}

func (m *MockAdminService) FetchAdminService(ctx context.Context, dto *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error) {
	args := m.Called(ctx, dto, secret)
	return args.Get(0).(*dtos.ResponseAdminDto), args.Error(1)
}

func (m *MockAdminService) CreateAdminService(ctx context.Context, req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*dtos.ResponseAdminSignupDto), args.Error(1)
}

func (m *MockAdminService) ListAdminsService(ctx context.Context) (*dtos.ResponseListAdminsDto, error) {
	args := m.Called(ctx)
	return args.Get(0).(*dtos.ResponseListAdminsDto), args.Error(1)
}

func (m *MockAdminService) DeleteAdminService(ctx context.Context, username string) (*dtos.ResponseDeleteAdminDto, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*dtos.ResponseDeleteAdminDto), args.Error(1)
}
