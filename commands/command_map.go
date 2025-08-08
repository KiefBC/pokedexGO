package commands

import "fmt"

const (
	mapBaseURL = "https://pokeapi.co/api/v2/location-area/"
)

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
func CommandGetMaps(cfg *Config, args ...string) error {
	url := mapBaseURL
	if cfg.NextURL != "" {
		url = cfg.NextURL
	}

	areaMaps, err := GetResponse[AreaMaps](url, cfg.Cache)
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
func CommandGetMapsBack(cfg *Config, args ...string) error {
	url := mapBaseURL
	if cfg.PreviousURL != "" {
		url = cfg.PreviousURL
	}

	areaMaps, err := GetResponse[AreaMaps](url, cfg.Cache)
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
