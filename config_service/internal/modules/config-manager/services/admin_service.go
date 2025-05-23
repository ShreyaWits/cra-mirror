package services

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/models"
	"nps-config-service/internal/modules/config-manager/repositories"
	"time"

	"github.com/golang-jwt/jwt"
)

type AdminService struct {
	Repo repositories.IConfigRepo
}

type IAdminService interface {
	FetchAdminService(dto *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error)
	CreateAdminService(req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error)
}

func NewAdminService(repo repositories.IConfigRepo) IAdminService {
	return &AdminService{Repo: repo}
}

func (s *AdminService) CreateAdminService(req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error) {

	admin := &models.Admin{
		UserName: req.Username,
		Password: req.Password,
	}
	createAdmin, err := s.Repo.CreateAdmin(admin)
	if err != nil {
		return nil, err
	}
	return &dtos.ResponseAdminSignupDto{
		Success: true,
		Message: "Admin created successfully",
		AdminId: createAdmin.UserName,
	}, nil
}
func (s *AdminService) FetchAdminService(req *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error) {
	// Create access token (1 hour expiry)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(),
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

	createAdmin, err := s.Repo.GetAdminByCredentials(admin.UserName, admin.Password)
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
