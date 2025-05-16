package constants

import "time"

const (
	SERVICE_NAME      = "messaging-service"
	CACCHE_TTL        = time.Hour * 240
	CACHE_SERVICE_URL = "host.docker.internal:50051"
)
