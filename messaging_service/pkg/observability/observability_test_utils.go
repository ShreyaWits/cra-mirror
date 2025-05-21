package observability

import (
	"strings"
)

// FormatEndpoint is an exported version of formatEndpoint for testing
// It removes any protocol prefix and ensures the endpoint is in the correct format
func FormatEndpoint(url string) string {
	// Remove http:// or https:// if present
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}
