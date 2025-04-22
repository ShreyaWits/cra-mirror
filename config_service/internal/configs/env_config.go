package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	EtcdEndpoint string
	APISecret    string
}

var AppConfig Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env not found, using default values")
	}
	AppConfig = Config{
		Port:         os.Getenv("PORT"),
		EtcdEndpoint: os.Getenv("ETCD_ENDPOINT"),
		APISecret:    os.Getenv("API_SECRET"),
	}
}
