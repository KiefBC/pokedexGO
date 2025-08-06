package commands

import "github.com/kiefbc/pokedexcli/internal/pokecache"

type Config struct {
	NextURL     string
	PreviousURL string
	Cache       *pokecache.Cache
}

type CliCommand struct {
	Name        string
	Description string
	Callback    func(*Config) error
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
	}
}
