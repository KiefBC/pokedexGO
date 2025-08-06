package commands

import (
	"fmt"
)

// CommandHelp displays the help message with all available commands and their descriptions.
// It prints a welcome message followed by usage information for each registered command.
func CommandHelp(cfg *Config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range GetCommands() {
		fmt.Printf("%s: %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println()
	return nil
}
