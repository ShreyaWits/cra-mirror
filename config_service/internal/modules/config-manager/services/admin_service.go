package services

import (
	"context"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/models"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/pkg/observability"
	"time"

	"github.com/golang-jwt/jwt"
)

type AdminService struct {
	Repo               repositories.IConfigRepo
	ObservabilityStack *observability.ObservabilityStack
}

type IAdminService interface {
	FetchAdminService(ctx context.Context, dto *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error)
	CreateAdminService(ctx context.Context, req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error)
}

func NewAdminService(repo repositories.IConfigRepo, ObservabilityStack *observability.ObservabilityStack) IAdminService {
	return &AdminService{Repo: repo, ObservabilityStack: ObservabilityStack}
}

func (s *AdminService) CreateAdminService(ctx context.Context,req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error) {

	admin := &models.Admin{
		UserName: req.Username,
		Password: req.Password,
	}
	createAdmin, err := s.Repo.CreateAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}
	return &dtos.ResponseAdminSignupDto{
		Success: true,
		Message: "Admin created successfully",
		AdminId: createAdmin.UserName,
	}, nil
}
func (s *AdminService) FetchAdminService(ctx context.Context, req *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error) {
	// Create access token (1 hour expiry)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	refreshSecret := secret // fallback to same secret if not set

	// Create refresh token (7 days expiry)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, err
	}

	admin := &models.Admin{
		UserName: req.Username,
		Password: req.Password,
	}

	createAdmin, err := s.Repo.GetAdminByCredentials(ctx, admin.UserName, admin.Password)
	if err != nil {
		return nil, err
	}

	// Return both tokens
	return &dtos.ResponseAdminDto{
		Success:      true,
		Message:      "Admin created successfully",
		AdminId:      createAdmin.UserName,
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
