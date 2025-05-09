package services

import (
	v1 "thirdparty_service/api/rest/v1/schemas"
	"thirdparty_service/internal/models"
	"thirdparty_service/internal/repositories"
	"thirdparty_service/utils"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req v1.CreateUserRequest) (v1.UserResponse, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return v1.UserResponse{}, err
	}

	user, err := s.repo.CreateUser(models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		return v1.UserResponse{}, err
	}

	return v1.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}
func (s *UserService) GetUserById(id string) (v1.UserResponse, error) {
	user, err := s.repo.GetUserById(id)
	if err != nil {
		return v1.UserResponse{}, err
	}
	return v1.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func (s *UserService) GetUserByEmail(email string) (v1.UserResponse, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return v1.UserResponse{}, err
	}

	return v1.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func (s *UserService) UpdateUser(req v1.UpdateUserRequest) error {
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
