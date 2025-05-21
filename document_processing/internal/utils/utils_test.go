package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveBytesToTempFile(t *testing.T) {
	// Create a temporary directory for testing
	tempTestDir, err := ioutil.TempDir("", "temp-test")
	if err != nil {
		t.Fatalf("Failed to create temp test directory: %v", err)
	}
	defer os.RemoveAll(tempTestDir) // Clean up the temporary directory

	// Change current working directory to the temporary test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tempTestDir);
	err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(oldWd) // Change back to the original directory

	tests := []struct {
		name     string
		data     []byte
		mimeType string
		expectExt string
		wantErr  bool
	}{
		{"png image", []byte("test png data"), "image/png", "png", false},
		{"jpeg image", []byte("test jpeg data"), "image/jpeg", "jpg", false},
		{"jpg image", []byte("test jpg data"), "image/jpg", "jpg", false},
		{"pdf document", []byte("test pdf data"), "application/pdf", "pdf", false},
		{"binary data", []byte("test binary data"), "application/octet-stream", "bin", false},
		{"empty data", []byte(""), "image/png", "png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := SaveBytesToTempFile(tt.data, tt.mimeType)

			if (err != nil) != tt.wantErr {
				t.Errorf("SaveBytesToTempFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the file was created
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("SaveBytesToTempFile() did not create file at %s", filePath)
				}

				// Verify the file content
				readData, err := ioutil.ReadFile(filePath)
				if err != nil {
					t.Errorf("SaveBytesToTempFile() failed to read created file: %v", err)
				}
				if string(readData) != string(tt.data) {
					t.Errorf("SaveBytesToTempFile() wrote incorrect data. Got %s, want %s", string(readData), string(tt.data))
				}

				// Verify the file extension
				ext := filepath.Ext(filePath)
				if strings.TrimPrefix(ext, ".") != tt.expectExt {
					t.Errorf("SaveBytesToTempFile() created file with incorrect extension. Got %s, want %s", ext, tt.expectExt)
				}

				// Clean up the created file
				defer os.Remove(filePath)
			}
		})
	}
}

func TestGetIdentityUserDetailsPrompt(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want string
	}{
		{
			name:   "empty fields",
			fields: []string{},
			want: `You are an AI system that extracts identity details from a document image.

Please find the following details from the document:


Return the details in the following JSON format without any code blocks, markdown formatting, or explanations:

If valid details are found:
{
}

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

Now, process the document and extract the required identity details.`,
		},
		{
			name:   "single field",
			fields: []string{"name"},
			want: `You are an AI system that extracts identity details from a document image.

Please find the following details from the document:
- name


Return the details in the following JSON format without any code blocks, markdown formatting, or explanations:

If valid details are found:
{
  "name": "<name or null>"
}

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

Now, process the document and extract the required identity details.`,
		},
		{
			name:   "multiple fields",
			fields: []string{"name", "dob", "address"},
			want: `You are an AI system that extracts identity details from a document image.

Please find the following details from the document:
- name
- dob
- address


Return the details in the following JSON format without any code blocks, markdown formatting, or explanations:

If valid details are found:
{
  "name": "<name or null>",
  "dob": "<dob or null>",
  "address": "<address or null>"
}

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

Now, process the document and extract the required identity details.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIdentityUserDetailsPrompt(tt.fields)
			if got != tt.want {
				t.Errorf("GetIdentityUserDetailsPrompt() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
} 


