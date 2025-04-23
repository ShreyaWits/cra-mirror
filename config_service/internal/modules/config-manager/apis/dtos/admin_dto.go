package dtos

type AdminDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseAdminDto struct {
	Success bool   `json:"success"`
	Token string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}