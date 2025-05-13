package services

import (
	"thirdparty_service/internal/models"
	"thirdparty_service/internal/modules/user/apis/dtos"
	repositories "thirdparty_service/internal/modules/user/repository"
	"thirdparty_service/internal/utils"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req dtos.CreateUserRequest) (dtos.UserResponse, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return dtos.UserResponse{}, err
	}

	user, err := s.repo.CreateUser(models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		return dtos.UserResponse{}, err
	}

	return dtos.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}
func (s *UserService) GetUserById(id string) (dtos.UserResponse, error) {
	user, err := s.repo.GetUserById(id)
	if err != nil {
		return dtos.UserResponse{}, err
	}
	return dtos.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func (s *UserService) GetUserByEmail(email string) (dtos.UserResponse, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return dtos.UserResponse{}, err
	}

	return dtos.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func (s *UserService) UpdateUser(req dtos.UpdateUserRequest) error {
	user := models.User{
		BaseModel: models.BaseModel{
			ID: req.ID,
		},
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	}
	return s.repo.UpdateUser(user)
}

func (s *UserService) DeleteUser(id string) error {
	return s.repo.DeleteUser(id)
}
