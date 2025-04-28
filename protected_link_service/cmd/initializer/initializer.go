package initializer

import (
	"protected_link/internal/common/utils"
	configEnv "protected_link/internal/configs"
	database "protected_link/pkg/redis"
)

func InitialSetup() (*configEnv.Config, *database.RedisConfig) {

	// Load messages from JSON
	utils.LoadMessages()

	// Load configuration
	cfg, err := configEnv.LoadConfig()

	// Connect to Redis
	db, eror := database.ConnectRedis(cfg)

	if err != nil {
		println("Error loading config:", err)
	}

	if eror != nil {
		println("Error db loading config:", err)
	}

	return cfg, db
}
