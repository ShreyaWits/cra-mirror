package http

// HTTPClientInterface defines the methods for making HTTP requests
type HttpClient interface {
	// Get makes a GET request to the specified URL and returns the response body
	Get(url string, headers map[string]string) ([]byte, error)

	// Post makes a POST request to the specified URL with the given payload and returns the response body
	Post(url string, payload interface{}, headers map[string]string) ([]byte, error)
}
