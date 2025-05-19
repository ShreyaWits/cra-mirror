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
	var client *minio.Client
	var err error

	// Retry MinIO client initialization
	maxRetries := 10
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		// Initialize MinIO client
		client, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false, // Use HTTP instead of HTTPS
		})
		if err == nil {
			// Client initialized successfully, now check bucket existence
			ctx := context.Background()
			var exists bool
			var bucketErr error

			// Retry bucket existence check
			for j := 0; j < maxRetries; j++ {
				exists, bucketErr = client.BucketExists(ctx, bucketName)
				if bucketErr == nil {
					// Bucket check successful, proceed
					if !exists {
						bucketErr = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
						if bucketErr != nil {
							return nil, fmt.Errorf("failed to create bucket after retries: %v", bucketErr)
						}
					}
					// Both client and bucket are ready
					repo := &MinioRepository{
						client:     client,
						bucketName: bucketName,
					}
					return repo, nil
				}

				fmt.Printf("MinIO bucket existence check attempt %d failed: %v. Retrying in %s...\n", j+1, bucketErr, retryDelay)
				time.Sleep(retryDelay)
			}
			// If bucket check failed after max retries
			return nil, fmt.Errorf("failed to check bucket existence after %d retries: %v", maxRetries, bucketErr)
		}

		fmt.Printf("MinIO client initialization attempt %d failed: %v. Retrying in %s...\n", i+1, err, retryDelay)
		time.Sleep(retryDelay)
	}

	// If client initialization failed after max retries
	return nil, fmt.Errorf("failed to create minio client after %d retries: %v", maxRetries, err)
}

// StoreFile stores a file in MinIO and returns its URL
func (r *MinioRepository) StoreFile(ctx context.Context, fileData []byte, fileType string) (string, error) {
	// Generate unique file name
	fileName := fmt.Sprintf("%s-%s", uuid.New().String(), time.Now().Format("20060102150405"))

	// Create a reader from the file data
	reader := bytes.NewReader(fileData)

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
