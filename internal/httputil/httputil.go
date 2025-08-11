package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kiefbc/pokedexcli/internal/pokecache"
)

const (
	DefaultTimeoutSeconds = 10
)

// NewClient creates a new HTTP client with the specified timeout
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// NewDefaultClient creates a new HTTP client with the default timeout
func NewDefaultClient() *http.Client {
	return NewClient(DefaultTimeoutSeconds * time.Second)
}

// GetResponse makes an HTTP GET request to the specified URL and parses the JSON response into the provided type T.
// It first checks the cache for existing data. If not found, it makes the HTTP request and caches the response.
// Returns the parsed response of type T and an error if the request fails, status is non-200, or JSON parsing fails.
func GetResponse[T any](url string, cache *pokecache.Cache, client *http.Client) (T, error) {
	var result T

	if cache == nil {
		return result, fmt.Errorf("cache cannot be nil")
	}

	if client == nil {
		client = NewDefaultClient()
	}

	// Check cache first
	if cached, found := cache.Get(url); found {
		err := json.Unmarshal(cached, &result)
		if err != nil {
			return result, fmt.Errorf("failed to unmarshal cached JSON: %w", err)
		}
		return result, nil
	}

	// Make HTTP request if not cached
	resp, err := client.Get(url)
	if err != nil {
		return result, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	// Cache the response
	err = cache.Add(url, body)
	if err != nil {
		return result, fmt.Errorf("failed to cache response: %w", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

// GetResponseWithDefault is a convenience wrapper that uses the default client
func GetResponseWithDefault[T any](url string, cache *pokecache.Cache) (T, error) {
	return GetResponse[T](url, cache, nil)
}
