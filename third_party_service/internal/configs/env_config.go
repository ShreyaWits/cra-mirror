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

var ThirdPartyConfig Config

// We will need it in fututre for configuration service to get env variables
func GetenvFromConfigService() {

}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env not found, using default values")
	}
	ThirdPartyConfig = Config{
		Port:         os.Getenv("PORT"),
		EtcdEndpoint: os.Getenv("ETCD_ENDPOINT"),
		APISecret:    os.Getenv("API_SECRET"),
	}

}
