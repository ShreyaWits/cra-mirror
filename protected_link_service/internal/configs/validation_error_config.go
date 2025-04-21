package configEnv

import (
	"encoding/json"
	"fmt"
	"os"
)

var ValidationErrors map[string]string

func LoadValidationErrors(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read error config file: %w", err)
	}
	if err := json.Unmarshal(data, &ValidationErrors); err != nil {
		return fmt.Errorf("failed to unmarshal error config: %w", err)
	}
	return nil
}

func GetValidationMessage(code string) string {
	if msg, ok := ValidationErrors[code]; ok {
		return msg
	}
	return "Unknown error code"
}
