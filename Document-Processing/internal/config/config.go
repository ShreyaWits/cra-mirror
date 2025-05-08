package config

type Config struct {
	ServerPort int
}

func NewConfig() *Config {
	return &Config{
		ServerPort: 50051,
	}
}
