package main

import (
	"bytes"
	"github.com/kiefbc/pokedexcli/commands"
	"github.com/kiefbc/pokedexcli/internal/pokecache"
	"io"
	"os"
	"testing"
	"time"
)

const (
	testCacheTimeout = 5 * time.Minute
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
		err := commands.CommandExit(&commands.Config{
			Cache: pokecache.NewCache(testCacheTimeout),
		})
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

		err := commands.CommandHelp(&commands.Config{
			Cache: pokecache.NewCache(testCacheTimeout),
		})

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

// TestCommandGetMaps tests the CommandGetMaps function to verify it handles
// API responses correctly and prints expected location area names.
func TestCommandGetMaps(t *testing.T) {
	cases := []struct {
		name             string
		initialNextURL   string
		expectedContains []string
	}{
		{
			name:           "first page request",
			initialNextURL: "",
			expectedContains: []string{
				"canalave-city-area",
				"eterna-city-area",
				"pastoria-city-area",
			},
		},
	}

	for _, c := range cases {
		// Create config
		cfg := &commands.Config{
			NextURL: c.initialNextURL,
			Cache:   pokecache.NewCache(testCacheTimeout),
		}

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call CommandGetMaps
		err := commands.CommandGetMaps(cfg)

		// Restore stdout
		w.Close()
		os.Stdout = old

		// Read captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actual := buf.String()

		// Check for errors
		if err != nil {
			t.Errorf("CommandGetMaps() returned an error: %v", err)
		}

		// Check that output contains expected location names
		for _, expected := range c.expectedContains {
			if !bytes.Contains([]byte(actual), []byte(expected)) {
				t.Errorf("CommandGetMaps() output missing expected string: %q\nGot: %q", expected, actual)
			}
		}
	}
}

// TestCommandGetMapsBack tests the CommandGetMapsBack function to verify it handles
// API responses correctly and prints expected location area names from previous page.
func TestCommandGetMapsBack(t *testing.T) {
	cases := []struct {
		name               string
		initialPreviousURL string
		expectedContains   []string
	}{
		{
			name:               "previous page request",
			initialPreviousURL: "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
			expectedContains: []string{
				"canalave-city-area",
				"eterna-city-area",
				"pastoria-city-area",
			},
		},
	}

	for _, c := range cases {
		// Create config
		cfg := &commands.Config{
			PreviousURL: c.initialPreviousURL,
			Cache:       pokecache.NewCache(testCacheTimeout),
		}

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call CommandGetMapsBack
		err := commands.CommandGetMapsBack(cfg)

		// Restore stdout
		w.Close()
		os.Stdout = old

		// Read captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actual := buf.String()

		// Check for errors
		if err != nil {
			t.Errorf("CommandGetMapsBack() returned an error: %v", err)
		}

		// Check that output contains expected location names
		for _, expected := range c.expectedContains {
			if !bytes.Contains([]byte(actual), []byte(expected)) {
				t.Errorf("CommandGetMapsBack() output missing expected string: %q\nGot: %q", expected, actual)
			}
		}
	}
}

// TestCommandExploreMap tests the CommandExploreMap function to verify it handles
// different scenarios including missing arguments, valid location areas, and API responses.
func TestCommandExploreMap(t *testing.T) {
	cases := []struct {
		name             string
		args             []string
		expectError      bool
		expectedContains []string
		errorContains    string
	}{
		{
			name:          "no arguments provided",
			args:          []string{},
			expectError:   true,
			errorContains: "explore command requires a location area name",
		},
		{
			name:        "valid location area",
			args:        []string{"canalave-city-area"},
			expectError: false,
			expectedContains: []string{
				"Exploring canalave-city-area...",
				"Found Pokemon:",
			},
		},
		{
			name:        "location with uppercase converted to lowercase",
			args:        []string{"CANALAVE-CITY-AREA"},
			expectError: false,
			expectedContains: []string{
				"Exploring canalave-city-area...",
				"Found Pokemon:",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create config
			cfg := &commands.Config{
				Cache: pokecache.NewCache(testCacheTimeout),
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call CommandExploreMap
			err := commands.CommandExploreMap(cfg, c.args...)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			actual := buf.String()

			// Check error expectation
			if c.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if c.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(c.errorContains)) {
					t.Errorf("Expected error to contain %q, got: %v", c.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("CommandExploreMap() returned unexpected error: %v", err)
				}
			}

			// Check expected output
			for _, expected := range c.expectedContains {
				if !bytes.Contains([]byte(actual), []byte(expected)) {
					t.Errorf("CommandExploreMap() output missing expected string: %q\nGot: %q", expected, actual)
				}
			}
		})
	}
}

// TestCommandCatchPokemon tests the CommandCatchPokemon function to verify it handles
// different scenarios including missing arguments, valid Pokemon, and duplicate catches.
func TestCommandCatchPokemon(t *testing.T) {
	cases := []struct {
		name             string
		args             []string
		existingPokedex  map[string]commands.Pokemon
		expectError      bool
		expectedContains []string
		errorContains    string
	}{
		{
			name:          "no arguments provided",
			args:          []string{},
			expectError:   true,
			errorContains: "catch command requires a Pokemon name",
		},
		{
			name:             "valid pokemon catch attempt",
			args:             []string{"pikachu"},
			existingPokedex:  make(map[string]commands.Pokemon),
			expectError:      false,
			expectedContains: []string{"Throwing a Pokeball at pikachu..."},
		},
		{
			name: "pokemon already caught",
			args: []string{"pikachu"},
			existingPokedex: map[string]commands.Pokemon{
				"pikachu": {Name: "pikachu", Height: 4, Weight: 60},
			},
			expectError:      false,
			expectedContains: []string{"pikachu is already in your Pokedex!"},
		},
		{
			name:             "uppercase pokemon name converted to lowercase",
			args:             []string{"PIKACHU"},
			existingPokedex:  make(map[string]commands.Pokemon),
			expectError:      false,
			expectedContains: []string{"Throwing a Pokeball at pikachu..."},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create config with existing Pokedex
			cfg := &commands.Config{
				Cache:   pokecache.NewCache(testCacheTimeout),
				Pokedex: c.existingPokedex,
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call CommandCatchPokemon
			err := commands.CommandCatchPokemon(cfg, c.args...)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			actual := buf.String()

			// Check error expectation
			if c.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if c.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(c.errorContains)) {
					t.Errorf("Expected error to contain %q, got: %v", c.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("CommandCatchPokemon() returned unexpected error: %v", err)
				}
			}

			// Check expected output
			for _, expected := range c.expectedContains {
				if !bytes.Contains([]byte(actual), []byte(expected)) {
					t.Errorf("CommandCatchPokemon() output missing expected string: %q\nGot: %q", expected, actual)
				}
			}
		})
	}
}

// TestCommandInspect tests the CommandInspect function to verify it handles
// different scenarios including missing arguments and Pokemon lookup.
func TestCommandInspect(t *testing.T) {
	cases := []struct {
		name             string
		args             []string
		pokedex          map[string]commands.Pokemon
		expectError      bool
		expectedContains []string
		errorContains    string
	}{
		{
			name:          "no arguments provided",
			args:          []string{},
			expectError:   true,
			errorContains: "inspect command requires a Pokemon name",
		},
		{
			name:             "pokemon not caught",
			args:             []string{"pikachu"},
			pokedex:          make(map[string]commands.Pokemon),
			expectError:      false,
			expectedContains: []string{"you have not caught that pokemon"},
		},
		{
			name: "valid pokemon inspection",
			args: []string{"pikachu"},
			pokedex: map[string]commands.Pokemon{
				"pikachu": {
					Name:           "pikachu",
					Height:         4,
					Weight:         60,
					BaseExperience: 112,
					Types:          []string{"electric"},
					ID:             25,
					Abilities:      []string{"static", "lightning-rod"},
				},
			},
			expectError: false,
			expectedContains: []string{
				"Pikachu (#25)", // Name will be bold but (#25) part is not
				"Height: 4 dm",
				"Weight: 60 hg",
				"Base Experience: 112",
				"Electric", // Type will be colored, so just check for the type name
			},
		},
		{
			name: "uppercase pokemon name converted to lowercase",
			args: []string{"PIKACHU"},
			pokedex: map[string]commands.Pokemon{
				"pikachu": {
					Name:           "pikachu",
					Height:         4,
					Weight:         60,
					BaseExperience: 112,
					Types:          []string{"electric"},
					ID:             25,
					Abilities:      []string{"static", "lightning-rod"},
				},
			},
			expectError: false,
			expectedContains: []string{
				"Pikachu (#25)", // Name will be bold but (#25) part is not
				"Height: 4 dm",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create config
			cfg := &commands.Config{
				Cache:   pokecache.NewCache(testCacheTimeout),
				Pokedex: c.pokedex,
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call CommandInspect
			err := commands.CommandInspect(cfg, c.args...)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			actual := buf.String()

			// Check error expectation
			if c.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if c.errorContains != "" && !bytes.Contains([]byte(err.Error()), []byte(c.errorContains)) {
					t.Errorf("Expected error to contain %q, got: %v", c.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("CommandInspect() returned unexpected error: %v", err)
				}
			}

			// Check expected output
			for _, expected := range c.expectedContains {
				if !bytes.Contains([]byte(actual), []byte(expected)) {
					t.Errorf("CommandInspect() output missing expected string: %q\nGot: %q", expected, actual)
				}
			}
		})
	}
}

// TestCommandPokedex tests the CommandPokedex function to verify it handles
// empty and populated Pokedex scenarios correctly.
func TestCommandPokedex(t *testing.T) {
	cases := []struct {
		name             string
		pokedex          map[string]commands.Pokemon
		expectedContains []string
	}{
		{
			name:             "empty pokedex",
			pokedex:          make(map[string]commands.Pokemon),
			expectedContains: []string{"Your Pokedex is empty."},
		},
		{
			name: "pokedex with single pokemon",
			pokedex: map[string]commands.Pokemon{
				"pikachu": {Name: "pikachu", Height: 4, Weight: 60},
			},
			expectedContains: []string{
				"Your Pokedex:",
				"  - pikachu",
			},
		},
		{
			name: "pokedex with multiple pokemon sorted alphabetically",
			pokedex: map[string]commands.Pokemon{
				"zubat":     {Name: "zubat", Height: 8, Weight: 75},
				"pikachu":   {Name: "pikachu", Height: 4, Weight: 60},
				"bulbasaur": {Name: "bulbasaur", Height: 7, Weight: 69},
			},
			expectedContains: []string{
				"Your Pokedex:",
				"  - bulbasaur",
				"  - pikachu",
				"  - zubat",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create config
			cfg := &commands.Config{
				Cache:   pokecache.NewCache(testCacheTimeout),
				Pokedex: c.pokedex,
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call CommandPokedex
			err := commands.CommandPokedex(cfg)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			actual := buf.String()

			// Check for errors
			if err != nil {
				t.Errorf("CommandPokedex() returned unexpected error: %v", err)
			}

			// Check expected output
			for _, expected := range c.expectedContains {
				if !bytes.Contains([]byte(actual), []byte(expected)) {
					t.Errorf("CommandPokedex() output missing expected string: %q\nGot: %q", expected, actual)
				}
			}
		})
	}
}
