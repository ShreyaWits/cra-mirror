package utils

import (
	"log"
	"os"
)

// ValidateEnvVars checks that required environment variables are set
func ValidateEnvVars(requiredVars []string) {
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("❌ Required environment variable %s is not set", v)
		}
	}
	log.Println("✅ All required environment variables are set")
}
