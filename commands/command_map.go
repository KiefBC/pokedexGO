package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func CommandGetMaps(cfg *Config) error {
	url := "https://pokeapi.co/api/v2/location-area/"
	if cfg.NextURL != "" {
		url = cfg.NextURL
	}
	areaMaps, err := getResponse(url)
	if err != nil {
		return fmt.Errorf("failed to get previous maps: %w", err)
	}

	// Update config with new pagination URLs
	cfg.NextURL = areaMaps.Next
	if areaMaps.Previous != nil {
		if prevStr, ok := areaMaps.Previous.(string); ok {
			cfg.PreviousURL = prevStr
		}
	}

	// Display the results
	for _, result := range areaMaps.Results {
		fmt.Printf("%s\n", result.Name)
	}

	return nil
}

func CommandGetMapsBack(cfg *Config) error {
	url := "https://pokeapi.co/api/v2/location-area/"
	if cfg.PreviousURL != "" {
		url = cfg.PreviousURL
	}

	areaMaps, err := getResponse(url)
	if err != nil {
		return fmt.Errorf("failed to get previous maps: %w", err)
	}

	// Update config with new pagination URLs
	cfg.NextURL = areaMaps.Next
	if areaMaps.Previous != nil {
		if prevStr, ok := areaMaps.Previous.(string); ok {
			cfg.PreviousURL = prevStr
		}
	}

	// Display the results
	for _, result := range areaMaps.Results {
		fmt.Printf("%s\n", result.Name)
	}

	return nil
}

func getResponse(url string) (AreaMaps, error) {
	resp, err := http.Get(url)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var areaMaps AreaMaps
	err = json.Unmarshal(body, &areaMaps)
	if err != nil {
		return AreaMaps{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return areaMaps, nil
}
