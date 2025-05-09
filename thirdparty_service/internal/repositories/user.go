package repositories

import (
	"thirdparty_service/internal/models"
)

type UserRepository interface {
	CreateUser(user models.User) (models.User, error)
	GetUserById(id string) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(id string) error
}

type PostgresUserRepository struct {
	// db *gorm.DB
}

func NewPostgresUserRepository() *PostgresUserRepository {
	return &PostgresUserRepository{}
}

func (r *PostgresUserRepository) CreateUser(user models.User) (models.User, error) {
	// err := r.db.Create(&user).Error
	return user, nil
}

func (r *PostgresUserRepository) GetUserById(id string) (models.User, error) {
	var user models.User
	// err := r.db.Where("id = ?", id).First(&user).Error
	return user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	// err := r.db.Where("email = ?", email).First(&user).Error
	return user, nil
}

func (r *PostgresUserRepository) UpdateUser(user models.User) error {
	return nil
}

func (r *PostgresUserRepository) DeleteUser(id string) error {
	// var user models.User
	// if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
	// 	return err
	// }
	return nil
}
