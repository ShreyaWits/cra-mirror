package dtos

type AdminSignupDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Secret string `json:"secret"`
}

type AdminLoginDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseAdminDto struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	AdminId string `json:"adminId"`
	Token string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type ResponseAdminSignupDto struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	AdminId string `json:"adminId"`
}