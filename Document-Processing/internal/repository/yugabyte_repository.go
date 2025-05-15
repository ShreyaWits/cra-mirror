package repository

import (
	"Document-Processing/internal/models"

	"gorm.io/gorm"
)

type DocumentDataRepository interface {
	CreateDocumentData(data *models.DocumentData) (*models.DocumentData, error)
	GetDocumentDataByID(id string) (*models.DocumentData, error)
}

type documentDataRepository struct {
	db *gorm.DB
}

func NewDocumentDataRepository(db *gorm.DB) DocumentDataRepository {
	return &documentDataRepository{db: db}
}

func (r *documentDataRepository) CreateDocumentData(data *models.DocumentData) (*models.DocumentData, error) {
	result := r.db.Create(data)
	if result.Error != nil {
		return nil, result.Error
	}
	return data, nil
}

func (r *documentDataRepository) GetDocumentDataByID(id string) (*models.DocumentData, error) {
	var document models.DocumentData
	result := r.db.First(&document, "Batch_id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &document, nil
}