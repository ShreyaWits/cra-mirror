package registry

var allowedSources = map[string]bool{
	"order_service":   true,
	"user_service":    true,
	"payment_service": true,
}

func IsRegisteredSource(source string) bool {
	return allowedSources[source]
}
