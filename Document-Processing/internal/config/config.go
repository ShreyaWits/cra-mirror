package config

type Config struct {
	ServerPort      int
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	GeminiAPIKey    string
}

func NewConfig() *Config {
	return &Config{
		ServerPort:      50051,
		MinioEndpoint:   "localhost:9000",
		MinioAccessKey:  "minioadmin",
		MinioSecretKey:  "minioadmin",
		MinioBucketName: "document-processing",
		GeminiAPIKey:    "AIzaSyB5HIbNjfAtbT_zIaAiypesAynqUyEIm_8",
	}
}
