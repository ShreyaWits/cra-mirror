package services

import (
	"context"
	"fmt"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/models"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/pkg/observability"
	"time"

	"github.com/golang-jwt/jwt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type AdminService struct {
	Repo               repositories.IConfigRepo
	ObservabilityStack *observability.ObservabilityStack
}

type IAdminService interface {
	FetchAdminService(ctx context.Context, dto *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error)
	CreateAdminService(ctx context.Context, req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error)
	ListAdminsService(ctx context.Context) (*dtos.ResponseListAdminsDto, error)
	DeleteAdminService(ctx context.Context, username string) (*dtos.ResponseDeleteAdminDto, error)
}

func NewAdminService(repo repositories.IConfigRepo, ObservabilityStack *observability.ObservabilityStack) IAdminService {
	if ObservabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}
	return &AdminService{Repo: repo, ObservabilityStack: ObservabilityStack}
}

func (s *AdminService) CreateAdminService(ctx context.Context, req *dtos.AdminSignupDto) (*dtos.ResponseAdminSignupDto, error) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "AdminService.CreateAdmin")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Creating new admin",
		"username", req.Username)

	span.SetAttributes(
		attribute.String("username", req.Username),
	)

	admin := &models.Admin{
		UserName: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	createAdmin, err := s.Repo.CreateAdmin(ctx, admin)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to create admin",
			"error", err,
			"username", req.Username)
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	span.SetStatus(codes.Ok, "Admin created successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Admin created successfully",
		"username", createAdmin.UserName,
		"admin_id", createAdmin.UserName)

	return &dtos.ResponseAdminSignupDto{
		Success: true,
		Message: "Admin created successfully",
		AdminId: createAdmin.UserName,
	}, nil
}

func (s *AdminService) FetchAdminService(ctx context.Context, req *dtos.AdminLoginDto, secret string) (*dtos.ResponseAdminDto, error) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "AdminService.FetchAdmin")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Fetching admin",
		"username", req.Username)

	span.SetAttributes(
		attribute.String("username", req.Username),
	)
	admin := &models.Admin{
		UserName: req.Username,
		Password: req.Password,
	}
	admin, err := s.Repo.GetAdminByCredentials(ctx, admin.UserName, admin.Password)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to get admin credentials"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get admin credentials",
			"error", err,
			"username", req.Username)
		return nil, fmt.Errorf("failed to get admin credentials: %w", err)
	}

	// Create access token (1 hour expiry)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role":     admin.Role,
		"username": admin.UserName,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to sign access token"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to sign access token",
			"error", err)
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshSecret := secret // fallback to same secret if not set

	// Create refresh token (7 days expiry)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":    true,
		"username": admin.UserName,
		"role":     admin.Role,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to sign refresh token"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to sign refresh token",
			"error", err)
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	span.SetStatus(codes.Ok, "Admin fetched successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Admin fetched successfully",
		"username", admin.UserName)

	// Return both tokens
	return &dtos.ResponseAdminDto{
		Success:      true,
		Message:      "Admin created successfully",
		AdminId:      admin.UserName,
		Role:         admin.Role,
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// convertToAdminDTO converts a models.Admin to dtos.Admin
func convertToAdminDTO(admin *models.Admin) *dtos.Admin {
	return &dtos.Admin{
		ID:        admin.ID.String(),
		UserName:  admin.UserName,
		Role:      admin.Role,
		CreatedAt: admin.CreatedAt.Format(time.RFC3339),
		UpdatedAt: admin.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *AdminService) ListAdminsService(ctx context.Context) (*dtos.ResponseListAdminsDto, error) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "AdminService.ListAdmins")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Listing all admin users")

	admins, err := s.Repo.ListAdmins(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to list admin users",
			"error", err)
		return nil, fmt.Errorf("failed to list admin users: %w", err)
	}

	// Convert models to DTOs
	adminDTOs := make([]*dtos.Admin, len(admins))
	for i, admin := range admins {
		adminDTOs[i] = convertToAdminDTO(admin)
	}

	span.SetStatus(codes.Ok, "Admin users listed successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Successfully listed admin users",
		"count", len(admins))

	return &dtos.ResponseListAdminsDto{
		Success: true,
		Message: "Admin users retrieved successfully",
		Admins:  adminDTOs,
	}, nil
}

// DeleteAdminService deletes an admin user
func (s *AdminService) DeleteAdminService(ctx context.Context, username string) (*dtos.ResponseDeleteAdminDto, error) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "AdminService.DeleteAdmin")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Deleting admin user",
		"username", username)

	span.SetAttributes(
		attribute.String("username", username),
	)

	err := s.Repo.DeleteAdmin(ctx, username)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", err.Error()))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to delete admin user",
			"error", err,
			"username", username)
		return nil, fmt.Errorf("failed to delete admin user: %w", err)
	}

	span.SetStatus(codes.Ok, "Admin user deleted successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Successfully deleted admin user",
		"username", username)

	return &dtos.ResponseDeleteAdminDto{
		Success:  true,
		Message:  "Admin user deleted successfully",
		Username: username,
	}, nil
}
