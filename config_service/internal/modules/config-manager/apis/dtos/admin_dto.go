package dtos

type AdminDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseAdminDto struct {
	Token string `json:"token"`
}