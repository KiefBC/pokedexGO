package commands

import (
	"fmt"
	"strings"
)

const (
	baseURL = "https://pokeapi.co/api/v2/location-area/"
)

type LocationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func CommandExploreMap(cfg *Config, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("explore command requires a location area name")
	}

	locationName := strings.ToLower(args[0])
	url := fmt.Sprintf("%s%s", baseURL, locationName)

	locationArea, err := GetResponse[LocationArea](url, cfg.Cache)
	if err != nil {
		return fmt.Errorf("failed to explore %s: %w", locationName, err)
	}

	fmt.Printf("Exploring %s...\n", locationName)
	fmt.Println("Found Pokemon:")

	if len(locationArea.PokemonEncounters) == 0 {
		fmt.Println("No Pokemon found in this area.")
		return nil
	}

	for _, encounter := range locationArea.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}
