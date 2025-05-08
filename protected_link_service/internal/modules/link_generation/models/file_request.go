package models

type FileRequestData struct {
	FileUrl  string `json:"file_url" validate:"required"`
	FileType string `json:"file_type" validate:"required"`
	Size     string `json:"size" validate:"required"`
}