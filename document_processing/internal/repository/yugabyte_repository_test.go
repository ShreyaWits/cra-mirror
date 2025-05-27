package repository_test

import (
	"document_processing/internal/models"
	"document_processing/internal/repository"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.DocumentData{})
	assert.NoError(t, err)

	return db
}

func TestDocumentDataRepository_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDocumentDataRepository(db)

	doc := &models.DocumentData{
		Batch_id:  "batch_123",
		Data:      "{\"key\":\"value\"}",
		Status:    "processed",
		Message:   "No errors",
		CreatedAt: time.Now(),
	}

	createdDoc, err := repo.CreateDocumentData(doc)
	assert.NoError(t, err)
	assert.NotNil(t, createdDoc)
	assert.Equal(t, "batch_123", createdDoc.Batch_id)
	assert.Equal(t, "processed", createdDoc.Status)

	fetchedDoc, err := repo.GetDocumentDataByID("batch_123")
	assert.NoError(t, err)
	assert.NotNil(t, fetchedDoc)
	assert.Equal(t, "batch_123", fetchedDoc.Batch_id)
	assert.Equal(t, "{\"key\":\"value\"}", fetchedDoc.Data)

	// Test fetching non-existent Batch_id returns error
	_, err = repo.GetDocumentDataByID("non_existent_id")
	assert.Error(t, err)
}

func TestDocumentDataRepository_GetDocumentDataByID_EmptyID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewDocumentDataRepository(db)

	_, err := repo.GetDocumentDataByID("")
	assert.Error(t, err)
}

func TestDocumentDataRepository_CreateDocumentData_Error(t *testing.T) {
	db := setupTestDB(t)

	// Close the DB to simulate failure on create
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	sqlDB.Close()

	repo := repository.NewDocumentDataRepository(db)

	doc := &models.DocumentData{
		Batch_id: "batch_error",
		Data:     "invalid data",
	}

	createdDoc, err := repo.CreateDocumentData(doc)
	assert.Error(t, err)
	assert.Nil(t, createdDoc)
}
