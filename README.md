# Pokedex CLI

## Description

A command-line Pokedex application built in Go that allows users to explore the Pokemon world through an interactive REPL interface. Browse location maps, discover Pokemon, and manage your collection directly from the terminal.

## Features

### Current Features

- **Interactive REPL Interface**: User-friendly command-line interface with `pokedex >` prompt
- **Location Map Browsing**: Explore Pokemon world location areas using the PokeAPI
  - `map` - Browse forward through location area maps
  - `mapb` - Navigate back through previous location area maps
- **Intelligent Caching**: HTTP responses cached with 5-minute TTL for improved performance
- **Help System**: Built-in help command to discover available functionality
- **Graceful Exit**: Clean application termination with goodbye message

### Planned Features

- **Pokemon Exploration**: Explore specific location areas and discover Pokemon inhabitants
- **Pokemon Battles**: Engage in battles with wild Pokemon
- **Pokemon Capture**: Catch and add Pokemon to your personal collection
- **Pokemon Management**: View, organize, and manage your captured Pokemon
- **Pokemon Care**: Feed and care for your Pokemon to keep them healthy

## Installation & Setup

### Prerequisites

- Go 1.24.2 or higher

### Clone and Build

1. Clone the repository:
```bash
git clone https://github.com/kiefbc/pokedexcli.git
cd pokedexcli
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o pokedexcli
```

4. Run the Pokedex:
```bash
./pokedexcli
```

## Usage

Once started, you'll see the Pokedex prompt:

```
pokedex > 
```

### Available Commands

- `help` - Display available commands and usage information
- `map` - Show the next page of location area maps
- `mapb` - Show the previous page of location area maps  
- `exit` - Exit the Pokedex application

### Example Session

```bash
$ ./pokedexcli
pokedex > help
Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex
map: Get a list of area maps
mapb: Go back to previous list of maps

pokedex > map
canalave-city-area
eterna-city-area
pastoria-city-area
...

pokedex > exit
Closing the Pokedex... Goodbye!
```

## Development

### Testing

Run the full test suite:
```bash
go test
```

Run tests with verbose output:
```bash
go test -v
```

Run a specific test:
```bash
go test -run TestFunctionName
```

### Code Formatting

Format the codebase:
```bash
go fmt ./...
```

### Dependencies

Update dependencies:
```bash
go mod tidy
```

## Architecture

This application follows a modular REPL-based architecture:

- **main.go**: Core REPL loop and input processing
- **commands/**: Command implementations and shared configuration
- **internal/pokecache/**: HTTP response caching system with automatic cleanup
- **poke_test.go**: Comprehensive test suite

The application integrates with the [PokeAPI](https://pokeapi.co/) to fetch real Pokemon world data and implements intelligent caching to minimize API calls and improve performance.

## Contributing

This project follows standard Go conventions. Please ensure all tests pass and code is formatted before submitting contributions.