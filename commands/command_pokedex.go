package commands

import (
	"fmt"
	"sort"
)

func CommandPokedex(cfg *Config, args ...string) error {
	if len(cfg.Pokedex) == 0 {
		fmt.Println("Your Pokedex is empty.")
		return nil
	}

	fmt.Println("Your Pokedex:")

	// Get pokemon names and sort them alphabetically
	names := make([]string, 0, len(cfg.Pokedex))
	for name := range cfg.Pokedex {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}
