package schemas

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,max=25"`
	LastName  string `json:"last_name" validate:"required,max=25"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=20"`
}

type UpdateUserRequest struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name" validate:"required,max=25"`
	LastName  string `json:"last_name" validate:"required,max=25"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=20"`
}

type UserResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
