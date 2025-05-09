package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// SaveBytesToTempFile saves a byte slice to a temporary file and returns the path.
func SaveBytesToTempFile(data []byte, mimeType string) (string, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	// Create temp directory path relative to the current directory
	tempDir := filepath.Join(currentDir, "temp")
	log.Printf("Creating temp directory at: %s", tempDir)

	// Ensure the temp folder exists
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		log.Printf("Error creating temp directory: %v", err)
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Determine file extension based on MIME type
	var ext string
	switch mimeType {
	case "image/png":
		ext = "png"
	case "image/jpeg", "image/jpg":
		ext = "jpg"
	case "application/pdf":
		ext = "pdf"
	default:
		ext = "bin"
	}

	// Create file path with timestamp
	fileName := fmt.Sprintf("file-%d.%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(tempDir, fileName)

	// Log the file path
	log.Printf("Saving file to: %s", filePath)

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("Error writing file: %v", err)
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("Successfully saved file to: %s", filePath)
	return filePath, nil
}
