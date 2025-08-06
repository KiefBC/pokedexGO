package commands

type CliCommand struct {
	Name        string
	Description string
	Callback    func() error
}

// GetCommands returns a map of all available CLI commands.
// Each command is mapped by its name and contains metadata and callback functions.
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
	}
}
