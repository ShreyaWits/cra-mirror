package repository

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioRepository struct {
	client     *minio.Client
	bucketName string
}

func NewMinioRepository(endpoint, accessKey, secretKey, bucketName string) (*MinioRepository, error) {
	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Use HTTP instead of HTTPS
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	// Create repository instance
	repo := &MinioRepository{
		client:     client,
		bucketName: bucketName,
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return repo, nil
}

// StoreFile stores a file in MinIO and returns its URL
func (r *MinioRepository) StoreFile(ctx context.Context, fileData []byte, fileType string) (string, error) {
	// Generate unique file name
	fileName := fmt.Sprintf("%s-%s", uuid.New().String(), time.Now().Format("20060102150405"))

	// Create a reader from the file data
	reader := bytes.NewReader(fileData)
	fmt.Print("jjsdjdjdjbdjddjndjdn", r.bucketName)

	// Upload file to MinIO
	_, err := r.client.PutObject(ctx, r.bucketName, fileName, reader, int64(len(fileData)), minio.PutObjectOptions{
		ContentType: fileType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Generate presigned URL (valid for 7 days)
	url, err := r.client.PresignedGetObject(ctx, r.bucketName, fileName, time.Hour*24*7, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return url.String(), nil
}

// GetFile retrieves a file from MinIO
func (r *MinioRepository) GetFile(ctx context.Context, fileName string) ([]byte, error) {
	object, err := r.client.GetObject(ctx, r.bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}
	defer object.Close()

	// Read file content
	fileInfo, err := object.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	buffer := make([]byte, fileInfo.Size)
	_, err = object.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	return buffer, nil
}

// DeleteFile deletes a file from MinIO using its URL
func (r *MinioRepository) DeleteFile(ctx context.Context, fileURL string) error {
	// Extract the object name from the URL
	// The URL format is typically: http://endpoint/bucket/object-name
	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid file URL format")
	}
	
	// Get the last part of the URL path and strip query parameters
	rawObjectName := parts[len(parts)-1]
	objectName := strings.Split(rawObjectName, "?")[0] // ✅ This removes query params
	
	// Remove the file from MinIO
	err := r.client.RemoveObject(ctx, r.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %v", err)
	}

	return nil
}
