package schemas

type CreateEmailRequest struct {
	FirstName string `json:"first_name" validate:"required,max=25"`
	LastName  string `json:"last_name" validate:"required,max=25"`
	Email     string `json:"email" validate:"required,email"`
	Body      string `json:"body" `
}

// need to implement
type CreateEmailResponse struct {
}
