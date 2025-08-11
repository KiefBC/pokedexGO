package commands

import (
	"fmt"
	"regexp"

	"github.com/kiefbc/pokedexcli/internal/httputil"
	"github.com/kiefbc/pokedexcli/internal/pokecache"
)

// Remove the duplicated HTTP client - now using shared utility

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
	Stats          []string
	// Enhanced fields for sprite support
	ID             int      `json:"id,omitempty"`
	Abilities      []string `json:"abilities,omitempty"`
	SpriteURL      string   `json:"sprite_url,omitempty"`
	SpriteShiny    string   `json:"sprite_shiny,omitempty"`
	SpriteOfficial string   `json:"sprite_official,omitempty"`
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

// GetResponse is a convenience wrapper for the shared HTTP utility
func GetResponse[T any](url string, cache *pokecache.Cache) (T, error) {
	return httputil.GetResponseWithDefault[T](url, cache)
}

// ValidatePokemonName validates Pokemon names for security across all commands.
//
// This function prevents various injection attacks by:
// - Limiting name length to prevent buffer overflows
// - Restricting to safe character sets (alphanumeric, hyphens, dots)
// - Preventing empty names that could cause lookup issues
//
// The validation allows legitimate Pokemon names like "mr-mime" and "flabebe"
// while blocking potentially malicious input that could be used for:
// - Path traversal attacks (../)
// - Command injection (; rm -rf)
// - SQL injection ('; DROP TABLE)
//
// Returns an error if the Pokemon name is invalid, nil if valid.
func ValidatePokemonName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("Pokemon name cannot be empty")
	}
	if len(name) > 50 {
		return fmt.Errorf("Pokemon name too long (max 50 characters)")
	}

	// Allow alphanumeric characters, hyphens, and dots (for some Pokemon names like Mr. Mime)
	if !regexp.MustCompile(`^[a-zA-Z0-9\-\.]+$`).MatchString(name) {
		return fmt.Errorf("Pokemon name contains invalid characters (only letters, numbers, hyphens, and dots allowed)")
	}

	return nil
}
