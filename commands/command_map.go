package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kiefbc/pokedexcli/internal/pokecache"
)

const (
	httpTimeoutSeconds = 10
	pokeAPIBaseURL     = "https://pokeapi.co/api/v2/location-area/"
)

// HTTP client with timeout
var httpClient = &http.Client{
	Timeout: httpTimeoutSeconds * time.Second,
}

type AreaMaps struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// CommandGetMaps fetches and displays the next page of location area maps from the PokeAPI.
// It updates the config with new pagination URLs for future navigation.
// Returns an error if the API request fails or response parsing fails.
func CommandGetMaps(cfg *Config) error {
	url := pokeAPIBaseURL
	if cfg.NextURL != "" {
		url = cfg.NextURL
	}

	areaMaps, err := getResponse(url, cfg.Cache)
	if err != nil {
		return fmt.Errorf("failed to get maps: %w", err)
	}

	cfg.NextURL = areaMaps.Next
	if areaMaps.Previous != nil {
		if prevStr, ok := areaMaps.Previous.(string); ok {
			cfg.PreviousURL = prevStr
		}
	}

	printMaps(areaMaps)

	return nil
}

// CommandGetMapsBack fetches and displays the previous page of location area maps from the PokeAPI.
// It updates the config with new pagination URLs for future navigation.
// Returns an error if the API request fails or response parsing fails.
func CommandGetMapsBack(cfg *Config) error {
	url := pokeAPIBaseURL
	if cfg.PreviousURL != "" {
		url = cfg.PreviousURL
	}

	areaMaps, err := getResponse(url, cfg.Cache)
	if err != nil {
		return fmt.Errorf("failed to get previous maps: %w", err)
	}

	cfg.NextURL = areaMaps.Next
	if areaMaps.Previous != nil {
		if prevStr, ok := areaMaps.Previous.(string); ok {
			cfg.PreviousURL = prevStr
		}
	}

	printMaps(areaMaps)

	return nil
}

// printMaps outputs the names of all location areas from the provided AreaMaps struct to stdout.
// Each area name is printed on a separate line. This function does not return any values.
func printMaps(areaMaps AreaMaps) {
	for _, result := range areaMaps.Results {
		fmt.Printf("%s\n", result.Name)
	}
}

// getResponse makes an HTTP GET request to the specified URL and parses the JSON response into an AreaMaps struct.
// It first checks the cache for existing data. If not found, it makes the HTTP request and caches the response.
// Returns the parsed AreaMaps struct and an error if the request fails, status is non-200, or JSON parsing fails.
func getResponse(url string, cache *pokecache.Cache) (AreaMaps, error) {
	// Check cache first
	if cached, found := cache.Get(url); found {
		var areaMaps AreaMaps
		err := json.Unmarshal(cached, &areaMaps)
		if err != nil {
			return AreaMaps{}, fmt.Errorf("failed to unmarshal cached JSON: %w", err)
		}
		return areaMaps, nil
	}

	// Make HTTP request if not cached
	resp, err := httpClient.Get(url)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to make request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return AreaMaps{}, fmt.Errorf("received non-200 response: %s", resp.Status)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Cache the response
	err = cache.Add(url, body)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to cache response: %w", err)
	}

	var areaMaps AreaMaps
	err = json.Unmarshal(body, &areaMaps)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return areaMaps, nil
}
