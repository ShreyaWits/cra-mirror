package dtos

type AdminSignupDto struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Secret   string `json:"secret" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=ADMIN VIEWER"`
}

type AdminLoginDto struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type ResponseAdminDto struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AdminId      string `json:"adminId"`
	Token        string `json:"token"`
	Role         string `json:"role"`
	RefreshToken string `json:"refreshToken"`
}

type ResponseAdminSignupDto struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	AdminId string `json:"adminId"`
}

type ResponseListAdminsDto struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Admins  []*Admin `json:"admins"`
}

type ResponseDeleteAdminDto struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

// Admin represents an admin user without sensitive information
type Admin struct {
	ID        string `json:"id"`
	UserName  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
