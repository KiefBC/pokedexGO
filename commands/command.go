package commands

import (
	"encoding/json"
	"fmt"
	"github.com/kiefbc/pokedexcli/internal/pokecache"
	"io"
	"net/http"
	"time"
)

const (
	httpTimeoutSeconds = 10
)

// HTTP client with timeout
var httpClient = &http.Client{
	Timeout: httpTimeoutSeconds * time.Second,
}

type Config struct {
	NextURL     string
	PreviousURL string
	Cache       *pokecache.Cache
	Pokedex     map[string]Pokemon
}

type Pokemon struct {
	Name           string
	Height         int
	Weight         int
	BaseExperience int
	Types          []string
}

type CliCommand struct {
	Name        string
	Description string
	Callback    func(*Config, ...string) error
}

// GetCommands returns a map of all available CLI commands.
// Each command is mapped by its name and contains metadata and callback functions.
// Returns a map where keys are command names (strings) and values are CliCommand structs.
func GetCommands() map[string]CliCommand {
	return map[string]CliCommand{
		"help": {
			Name:        "help",
			Description: "Displays a help message",
			Callback:    CommandHelp,
		},
		"exit": {
			Name:        "exit",
			Description: "Exit the Pokedex",
			Callback:    CommandExit,
		},
		"map": {
			Name:        "map",
			Description: "Get a list of area maps",
			Callback:    CommandGetMaps,
		},
		"mapb": {
			Name:        "mapb",
			Description: "Go back to previous list of maps",
			Callback:    CommandGetMapsBack,
		},
		"explore": {
			Name:        "explore",
			Description: "Explore a specific area map",
			Callback:    CommandExploreMap,
		},
		"catch": {
			Name:        "catch",
			Description: "Catch a specific Pokemon",
			Callback:    CommandCatchPokemon,
		},
		"inspect": {
			Name:        "inspect",
			Description: "View details of a caught Pokemon",
			Callback:    CommandInspect,
		},
		"pokedex": {
			Name:        "pokedex",
			Description: "View all caught Pokemon",
			Callback:    CommandPokedex,
		},
	}
}

// GetResponse makes an HTTP GET request to the specified URL and parses the JSON response into the provided type T.
// It first checks the cache for existing data. If not found, it makes the HTTP request and caches the response.
// Returns the parsed response of type T and an error if the request fails, status is non-200, or JSON parsing fails.
func GetResponse[T any](url string, cache *pokecache.Cache) (T, error) {
	var result T

	// Check cache first
	if cached, found := cache.Get(url); found {
		err := json.Unmarshal(cached, &result)
		if err != nil {
			return result, fmt.Errorf("failed to unmarshal cached JSON: %w", err)
		}
		return result, nil
	}

	// Make HTTP request if not cached
	resp, err := httpClient.Get(url)
	if err != nil {
		return result, fmt.Errorf("failed to make request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("received non-200 response: %s", resp.Status)
	}
	defer resp.Body.Close()

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
