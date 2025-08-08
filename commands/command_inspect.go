package commands

import (
	"fmt"
	"strings"
)

func CommandInspect(cfg *Config, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("inspect command requires a pokemon name")
	}

	pokemonName := strings.ToLower(args[0])
	
	pokemon, exists := cfg.Pokedex[pokemonName]
	if !exists {
		fmt.Printf("you have not caught that pokemon\n")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Base experience: %d\n", pokemon.BaseExperience)
	fmt.Println("Types:")
	for _, pokemonType := range pokemon.Types {
		fmt.Printf("  - %s\n", pokemonType)
	}

	return nil
}