package repository_test

import (
	"Document-Processing/internal/repository"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Run minio container command
// docker run -d -p 9000:9000 -p 9001:9001 \
//   --name minio \
//   -e "MINIO_ROOT_USER=minioadmin" \
//   -e "MINIO_ROOT_PASSWORD=minioadmin" \
//   quay.io/minio/minio server /data --console-address ":9001"

func TestMinioRepository(t *testing.T) {
	ctx := context.Background()
	repo, err := repository.NewMinioRepository("localhost:9000", "minioadmin", "minioadmin", "test-bucket")
	assert.NoError(t, err)

	// Test StoreFile
	fileData := []byte("this is a test file")
	url, err := repo.StoreFile(ctx, fileData, "text/plain")
	assert.NoError(t, err)
	assert.Contains(t, url, "http://")

	// Test GetFile
	parts := strings.Split(url, "/")
	fileName := parts[len(parts)-1]
	fileName = strings.Split(fileName, "?")[0]

	downloaded, err := repo.GetFile(ctx, fileName)
	assert.NoError(t, err)
	assert.Equal(t, fileData, downloaded)

	// Test DeleteFile
	err = repo.DeleteFile(ctx, url)
	assert.NoError(t, err)
}

func TestNewMinioRepository_InvalidHost(t *testing.T) {
	_, err := repository.NewMinioRepository("invalid:1234", "minioadmin", "minioadmin", "test-bucket")
	assert.Error(t, err)
}

func TestDeleteFile_InvalidURL(t *testing.T) {
	repo, err := repository.NewMinioRepository("localhost:9000", "minioadmin", "minioadmin", "test-bucket")
	assert.NoError(t, err)

	// Invalid format (too short)
	err = repo.DeleteFile(context.Background(), "http://short-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file URL format")
}

func TestDeleteFile_NonExistentObject(t *testing.T) {
	repo, err := repository.NewMinioRepository("localhost:9000", "minioadmin", "minioadmin", "test-bucket")
	assert.NoError(t, err)

	// Random URL that points to a non-existent object
	url := "http://localhost:9000/test-bucket/non-existent-object"
	err = repo.DeleteFile(context.Background(), url)
	assert.Error(t, err)
}
