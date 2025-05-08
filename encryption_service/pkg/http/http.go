package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// HttpClient is a struct that provides methods for making Http requests
type HttpClientStruct struct {
	client *http.Client
}

// NewHttpClient creates a new instance of HttpClient
func NewHttpClient() HttpClient {
	return &HttpClientStruct{
		client: http.DefaultClient,
	}
}

// Get makes a GET request to the specified URL and returns the response body
func (h *HttpClientStruct) Get(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Use the Http client to make the request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch data: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Post makes a POST request to the specified URL with the given payload and returns the response body
func (h *HttpClientStruct) Post(url string, payload interface{}, headers map[string]string) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Use the Http client to make the request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to post data: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}
