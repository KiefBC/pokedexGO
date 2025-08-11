package main

import (
	"bufio"
	"fmt"
	"github.com/kiefbc/pokedexcli/commands"
	"github.com/kiefbc/pokedexcli/internal/pokecache"
	"os"
	"strings"
	"time"
)

const (
	maxCommandLength   = 50
	cacheTimeoutLength = 5 * time.Minute
)

// main starts the Pokedex CLI application and enters the REPL loop.
// It continuously prompts for user input, processes commands, and executes them.
// This function does not return - it runs until the program exits via a command.
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cache := pokecache.NewCache(cacheTimeoutLength)

	cfg := &commands.Config{
		Cache:   cache,
		Pokedex: make(map[string]commands.Pokemon),
	}

	for {
		fmt.Print("pokedex > ")
		scanner.Scan()
		userInput := cleanInput(scanner.Text())
		if len(userInput) == 0 {
			continue
		}

		if len(userInput[0]) > maxCommandLength {
			fmt.Println("Command too long")
			continue
		}

		command := userInput[0]

		if cmd, exists := commands.GetCommands()[command]; exists {
			args := userInput[1:]
			err := cmd.Callback(cfg, args...)
			if err != nil {
				fmt.Printf("Error executing command '%s': %v\n", command, err)
			}
		} else {
			fmt.Printf("Unknown command\n")
		}
	}
}

// cleanInput takes a raw text string and returns a cleaned slice of strings.
// It converts the input to lowercase and splits it by whitespace.
// Returns a slice of strings where each element is a whitespace-separated word from the input.
func cleanInput(text string) []string {
	var input []string
	input = strings.Fields(strings.ToLower(text))
	return input
}
