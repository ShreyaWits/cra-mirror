package config_test

import (
	"messaging_service/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig with default values", func(t *testing.T) {
		// Clear relevant environment variables to ensure defaults are used
		os.Clearenv()

		_, err := config.LoadConfig()
		assert.NoError(t, err)
	})

	t.Run("LoadConfig with environment variables", func(t *testing.T) {
		os.Setenv("GRPC_PORT", "60000")
		os.Setenv("ENCRIPTION_SECRET", "custom_secret")
		os.Setenv("DB_HOST", "custom_host")

		_, err := config.LoadConfig()
		assert.NoError(t, err)

		// Clean up
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("ENCRIPTION_SECRET")
		os.Unsetenv("DB_HOST")
	})
}
