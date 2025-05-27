package configEnv

type ImmutableConfig struct {
	ConfigServiceUrl   string `env:"CONFIG_SERVICE_URL" json:"configServiceUrl"`
	ConfigServiceToken string `env:"CONFIG_SERVICE_TOKEN" json:"configServiceToken"`
	Environment        string `env:"ENVIRONMENT" json:"environment"`
	CacheSrvAddr       string `env:"CACHE_SERVICE_ADDR" json:"cacheServiceAddr"`
	TemplateServiceGRPCPort string `env:"TEMPLATE_SERVICE_GRPC_PORT" json:"templateServiceGRPCPort"`
	TemplateServiceRestPort  string `env:"TEMPLATE_SERVICE_REST_PORT" json:"templateServiceRestPort"`
}

type Config struct {
	// Server settings
	HttpPort string `json:"REST_PORT"`
	GRPCPort string `json:"GRPC_PORT"`

	// YugabyteDB settings
	YugabyteDBHost     string `json:"YUGABYTE_DATABASE_HOST"`
	YugabyteDBPort     string `json:"YUGABYTE_DATABASE_PORT"`
	YugabyteDBUser     string `json:"YUGABYTE_DATABASE_USER"`
	YugabyteDBPassword string `json:"YUGABYTE_DATABASE_PASSWORD"`
	YugabyteDBName     string `json:"TEMPLATE_SERVICE_YUGABYTE_DATABASE_NAME"`

	// Cache and other settings
	CacheUrl         string `json:"CACHE_URL"`
	CacheTtl         string `json:"CACHE_TTL"`
	ObservabilityUrl string `json:"OBSERVABILITY_URL"`
	Environment      string `json:"ENVIRONMENT"`
}

type ConfigServiceResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       Config `json:"data"`
}

type ConfigServiceWebhookData struct {
	Environment string `json:"environment"`
	Method      string `json:"method"`
	ServiceName string `json:"serviceName"`
	Values      Config `json:"values"`
}

type RegisterWebHookRequest struct {
	URL         string `json:"url"`
	Environment string `json:"environment"`
	ServiceName string `json:"serviceName"`
	Method      string `json:"Method"`
}
