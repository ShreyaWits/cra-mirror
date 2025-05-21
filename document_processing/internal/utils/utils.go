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

func GetIdentityUserDetailsPrompt(fields []string) string {
	prompt := `You are an AI system that extracts identity details from a document image.

Please find the following details from the document:
`

	for _, field := range fields {
		prompt += fmt.Sprintf("- %s\n", field)
	}

	prompt += `

Return the details in the following JSON format without any code blocks, markdown formatting, or explanations:

If valid details are found:
{
`

	for i, field := range fields {
		comma := ","
		if i == len(fields)-1 {
			comma = ""
		}
		prompt += fmt.Sprintf("  \"%s\": \"<%s or null>\"%s\n", field, field, comma)
	}

	prompt += `}

If no valid identity details are found:
{
  "error": "document is not valid"
}

IMPORTANT INSTRUCTIONS:
1. Your response must ONLY contain the raw JSON object
2. Do not include 'json' tags
3. Do not include any explanations before or after the JSON
4. Do not use escape characters for newlines (\n) within the JSON values
5. Format the JSON on multiple lines with proper indentation
6. Keep address or any text field as a single line without line breaks
7. Return field values exactly as they appear in the document

Now, process the document and extract the required identity details.`

	return prompt
}
