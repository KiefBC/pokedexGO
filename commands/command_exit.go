package commands

import (
	"fmt"
	"os"
)

type Exiter interface {
	Exit(code int)
}

type OSExiter struct{}

// Exit terminates the program with the specified exit code using os.Exit.
func (o OSExiter) Exit(code int) {
	os.Exit(code)
}

var exiter Exiter = OSExiter{}

// GetExiter returns the current exiter implementation used for program termination.
func GetExiter() Exiter {
	return exiter
}

// SetExiter sets the exiter implementation for program termination.
// This is primarily used for testing to inject mock exiters.
func SetExiter(e Exiter) {
	exiter = e
}

// CommandExit handles the exit command by displaying a goodbye message and terminating the application.
// It prints a farewell message and calls the configured exiter with status code 0.
func CommandExit(cfg *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	exiter.Exit(0)
	return nil
}
