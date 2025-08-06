package main

import (
	"bufio"
	"fmt"
	"github.com/kiefbc/pokedexcli/commands"
	"os"
	"strings"
)

const (
	maxCommandLength = 50
)

// main starts the Pokedex CLI application and enters the REPL loop.
// It continuously prompts for user input, processes commands, and executes them.
// This function does not return - it runs until the program exits via a command.
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := &commands.Config{}

	for {
		fmt.Print("pokedex > ")
		scanner.Scan()
		userInput := cleanInput(scanner.Text())
		if len(userInput) == 0 {
			continue
		}
		
		// Input validation
		if len(userInput[0]) > maxCommandLength {
			fmt.Println("Command too long")
			continue
		}
		
		command := userInput[0]
		// fmt.Printf("Your command was: %s\n", command)

		if cmd, exists := commands.GetCommands()[command]; exists {
			err := cmd.Callback(cfg)
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
