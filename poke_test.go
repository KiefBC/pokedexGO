package main

import (
	"bytes"
	"github.com/kiefbc/pokedexcli/commands"
	"io"
	"os"
	"testing"
)

// TestCleanInput tests the cleanInput function with various input scenarios.
// It verifies that input is properly cleaned, lowercased, and split into fields.
func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  hello, world!  ",
			expected: []string{"hello,", "world!"},
		},
		{
			input:    "  hello, world!  how are you?  ",
			expected: []string{"hello,", "world!", "how", "are", "you?"},
		},
		// add more cases here
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q) = %v; want %v", c.input, actual, c.expected)
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%q)[%d] = %q; want %q", c.input, i, word, expectedWord)
				return
			}
		}
	}
}

type MockExiter struct {
	ExitCode int
	Called   bool
}

// Exit is a mock implementation that records the exit code and call status
// instead of actually terminating the program.
func (m *MockExiter) Exit(code int) {
	m.ExitCode = code
	m.Called = true
}

// TestCommandExit tests the CommandExit function to ensure it prints the correct
// farewell message and calls the exiter with the expected exit code.
func TestCommandExit(t *testing.T) {
	cases := []struct {
		input        string
		expected     string
		expectedCode int
	}{
		{
			input:        "exit",
			expected:     "Closing the Pokedex... Goodbye!\n",
			expectedCode: 0,
		},
		// add more cases
	}

	for _, c := range cases {
		// Setup mock exiter
		mockExiter := &MockExiter{}
		originalExiter := commands.GetExiter()
		commands.SetExiter(mockExiter)
		defer commands.SetExiter(originalExiter)

		// Capture stdout to get the actual output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call the actual CommandExit function
		err := commands.CommandExit(&commands.Config{})
		// Restore stdout
		w.Close()
		os.Stdout = old

		// Read the captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actual := buf.String()

		// Check for errors
		if err != nil {
			t.Errorf("CommandExit() returned an error: %v", err)
		}

		// Compare actual vs expected output
		if actual != c.expected {
			t.Errorf("input: %q, got: %q, want: %q", c.input, actual, c.expected)
		}

		// Verify exit was called with correct code
		if !mockExiter.Called {
			t.Errorf("Expected Exit() to be called")
		}
		if mockExiter.ExitCode != c.expectedCode {
			t.Errorf("Expected exit code %d, got %d", c.expectedCode, mockExiter.ExitCode)
		}
	}
}

// TestCommandHelp tests the CommandHelp function to verify it displays
// the welcome message and all expected command information.
func TestCommandHelp(t *testing.T) {
	cases := []struct {
		input            string
		expectedContains []string
	}{
		{
			input: "help",
			expectedContains: []string{
				"Welcome to the Pokedex!",
				"Usage:",
				"help: Displays a help message",
				"exit: Exit the Pokedex",
			},
		},
		// add more cases
	}

	for _, c := range cases {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := commands.CommandHelp(&commands.Config{})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		actual := buf.String()

		if err != nil {
			t.Errorf("commandHelp() returned an error: %v", err)
		}

		for _, expected := range c.expectedContains {
			if !bytes.Contains([]byte(actual), []byte(expected)) {
				t.Errorf("commandHelp() output missing expected string: %q\nGot: %q", expected, actual)
			}
		}
	}
}
