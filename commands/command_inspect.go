package commands

import (
	"bytes"
	"fmt"
	"image"
	imgcolor "image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unsafe"

	"github.com/disintegration/imaging"
	"github.com/fatih/color"
	"github.com/kiefbc/pokedexcli/internal/sprites"
	"github.com/qeesung/image2ascii/convert"
)

// Display layout constants
const (
	asciiWidth       = 80
	asciiHeight      = 40
	infoBoxWidth     = 43  // 43 characters wide inside the box
	infoBoxPadding   = 42  // 42 spaces for padding (43 - 1 for content)
	sideSpacing      = 85  // spacing for side-by-side display
	minTerminalWidth = 130 // minimum width needed for proper ASCII art display
)

// getVisualLength calculates the visual length of text excluding ANSI color codes
func getVisualLength(text string) int {
	// Strip ANSI color codes for accurate length calculation
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return len(ansiRegex.ReplaceAllString(text, ""))
}

// getTerminalWidth returns the terminal width, or 0 if unable to determine
func getTerminalWidth() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(os.Stdout.Fd()),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 0 // Unable to determine terminal size
	}
	return int(ws.Col)
}

// CommandInspect displays detailed Pokemon information with beautiful ASCII art sprites.
//
// This enhanced command creates a stunning visual presentation combining:
// - High-quality ASCII art sprites (80x40 resolution)
// - Type-based color schemes for visual appeal
// - Neofetch-style side-by-side layout
// - Comprehensive Pokemon stats with visual progress bars
//
// The system uses smart caching - sprites are downloaded once and cached locally
// for instant display on subsequent inspections. No configuration needed.
//
// Usage: inspect <pokemon_name>
// Example: inspect pikachu
func CommandInspect(cfg *Config, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("inspect command requires a Pokemon name")
	}

	pokemonName := strings.ToLower(args[0])

	pokemon, exists := cfg.Pokedex[pokemonName]
	if !exists {
		fmt.Printf("you have not caught that pokemon\n")
		return nil
	}

	// Check terminal width for ASCII art display
	// Only switch to text-only if we can detect width AND it's narrow
	terminalWidth := getTerminalWidth()
	if terminalWidth > 0 && terminalWidth < minTerminalWidth {
		// Terminal too narrow - show text-only display
		displayPokemonTextOnly(pokemon)
		fmt.Printf("\n%s\n",
			color.New(color.FgYellow).Sprint("💡 Terminal too narrow for ASCII art. Resize to at least 130 characters wide to see Pokemon sprite!"))
		return nil
	}
	// Default to ASCII art mode if width unknown (like during tests) or wide enough

	// Try to get colorblock art from sprite
	asciiArt := getColorblockArt(pokemon)

	// Create the full display with ASCII art
	displayPokemon(pokemon, asciiArt)

	return nil
}

// getASCIIArt downloads sprite and converts to ASCII art using a simple, direct approach.
// No complex configuration - uses optimal settings for high-quality 80x40 ASCII art.
// Falls back to a simple Pokemon ball if sprite unavailable.
func getASCIIArt(pokemon Pokemon) []string {
	// Try official artwork first for best quality, fallback to regular sprite
	spriteURL := pokemon.SpriteOfficial
	if spriteURL == "" {
		spriteURL = pokemon.SpriteURL
	}

	if spriteURL == "" {
		return getFallbackASCII()
	}

	// Download with simple caching
	imageData, err := sprites.DownloadAndCacheSprite(spriteURL)
	if err != nil {
		return getFallbackASCII()
	}

	// Convert to ASCII art
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return getFallbackASCII()
	}

	// Resize for better ASCII conversion
	resized := imaging.Resize(img, 240, 120, imaging.Lanczos)

	// Convert to ASCII with high quality settings
	options := convert.DefaultOptions
	options.FixedWidth = asciiWidth
	options.FixedHeight = asciiHeight
	options.Colored = true
	options.Reversed = false

	converter := convert.NewImageConverter()
	asciiString := converter.Image2ASCIIString(resized, &options)

	return strings.Split(asciiString, "\n")
}

// getFallbackASCII provides a simple Pokemon ball ASCII art when sprites aren't available.
// This ensures users always get a visual representation, even without network access.
func getFallbackASCII() []string {
	return strings.Split(`
    ╭─────────╮
  ╱           ╲
 ╱    ┌─────┐  ╲
╱     │  ○  │   ╲
│ ────┼─────┼──── │
│     │     │     │
╲     └─────┘    ╱
 ╲               ╱
  ╲_____________╱
   Pokemon Ball`, "\n")
}

// getColorblockArt converts Pokemon sprites to high-quality colorblock art using Unicode half-blocks.
// This provides 2x higher vertical resolution than traditional block rendering by using the ▄ character
// with background color for top pixel and foreground color for bottom pixel.
func getColorblockArt(pokemon Pokemon) []string {
	// Try official artwork first for best quality, fallback to regular sprite
	spriteURL := pokemon.SpriteOfficial
	if spriteURL == "" {
		spriteURL = pokemon.SpriteURL
	}

	if spriteURL == "" {
		return getFallbackASCII()
	}

	// Download with simple caching
	imageData, err := sprites.DownloadAndCacheSprite(spriteURL)
	if err != nil {
		return getFallbackASCII()
	}

	// Convert to colorblock art
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return getFallbackASCII()
	}

	// Resize for colorblock conversion - maintain aspect ratio for 80x40 output
	// Since we use half-blocks, we need 80x80 pixels for 80x40 display
	resized := imaging.Resize(img, 80, 80, imaging.Lanczos)

	return convertToColorblocks(resized)
}

// convertToColorblocks converts an image to colorblock representation using Unicode half-blocks.
// Each character represents 2 vertical pixels using ▄ with background (top) and foreground (bottom) colors.
// Handles transparent pixels by skipping them or using space characters.
func convertToColorblocks(img image.Image) []string {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Detect background color from edge pixels
	backgroundColor := detectBackgroundColor(img)

	// Each character represents 2 vertical pixels, so we get height/2 rows
	rows := make([]string, 0, height/2)

	for y := 0; y < height; y += 2 {
		var line strings.Builder

		for x := 0; x < width; x++ {
			// Get top and bottom pixel colors
			topColor := img.At(x, y)

			// Handle case where we're at the last row and it's odd
			var bottomColor imgcolor.Color
			if y+1 < height {
				bottomColor = img.At(x, y+1)
			} else {
				bottomColor = topColor // Use same color if no bottom pixel
			}

			// Convert to ANSI color codes
			topAnsi := rgbToAnsi(topColor)
			bottomAnsi := rgbToAnsi(bottomColor)

			// Handle transparent pixels
			if topAnsi == -1 && bottomAnsi == -1 {
				// Both pixels transparent - use space
				line.WriteString(" ")
			} else if topAnsi == -1 {
				// Top pixel transparent, bottom visible - treat as background
				if isBackgroundColor(bottomColor, backgroundColor) {
					line.WriteString(" ")
				} else {
					line.WriteString(fmt.Sprintf("\x1b[38;5;%dm▄\x1b[0m", bottomAnsi))
				}
			} else if bottomAnsi == -1 {
				// Bottom pixel transparent, top visible - treat as background
				if isBackgroundColor(topColor, backgroundColor) {
					line.WriteString(" ")
				} else {
					line.WriteString(fmt.Sprintf("\x1b[48;5;%dm \x1b[0m", topAnsi))
				}
			} else {
				// Both pixels visible - check for background colors
				topIsBackground := isBackgroundColor(topColor, backgroundColor)
				bottomIsBackground := isBackgroundColor(bottomColor, backgroundColor)

				if topIsBackground && bottomIsBackground {
					// Both are background - use space
					line.WriteString(" ")
				} else if topIsBackground {
					// Top is background, bottom is visible
					line.WriteString(fmt.Sprintf("\x1b[38;5;%dm▄\x1b[0m", bottomAnsi))
				} else if bottomIsBackground {
					// Bottom is background, top is visible
					line.WriteString(fmt.Sprintf("\x1b[48;5;%dm \x1b[0m", topAnsi))
				} else {
					// Both are foreground colors - normal block
					line.WriteString(fmt.Sprintf("\x1b[48;5;%dm\x1b[38;5;%dm▄\x1b[0m", topAnsi, bottomAnsi))
				}
			}
		}

		rows = append(rows, line.String())
	}

	return rows
}

// rgbToAnsi converts RGB color to the nearest ANSI 256-color code.
// Uses a simplified approach that maps RGB values to the 6x6x6 color cube (colors 16-231)
// plus grayscale (colors 232-255) for better terminal compatibility.
// Returns -1 for transparent or semi-transparent pixels.
func rgbToAnsi(c imgcolor.Color) int {
	r, g, b, a := c.RGBA()

	// Handle transparency - return special value for transparent pixels
	// Check for both fully transparent (a=0) and semi-transparent (a < threshold)
	alpha8 := uint8(a >> 8)
	if alpha8 < 128 { // Semi-transparent threshold
		return -1 // Special value indicating transparency
	}

	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)

	// Check if it's grayscale (or close to it)
	// Reduced tolerance to avoid catching near-transparent pixels
	maxDiff := uint8(10) // Reduced tolerance for considering it grayscale
	avgColor := (int(r8) + int(g8) + int(b8)) / 3

	if abs(int(r8)-avgColor) <= int(maxDiff) &&
		abs(int(g8)-avgColor) <= int(maxDiff) &&
		abs(int(b8)-avgColor) <= int(maxDiff) {
		// Map to grayscale range (232-255)
		// 24 levels of gray from dark to light
		grayLevel := int(float64(avgColor) / 255.0 * 23)
		return 232 + grayLevel
	}

	// Map to 6x6x6 color cube (colors 16-231)
	// Each component has 6 levels: 0, 95, 135, 175, 215, 255
	rLevel := colorToLevel(r8)
	gLevel := colorToLevel(g8)
	bLevel := colorToLevel(b8)

	return 16 + 36*rLevel + 6*gLevel + bLevel
}

// colorToLevel maps an 8-bit color value to a 6-level index for the ANSI color cube
func colorToLevel(value uint8) int {
	// The 6 levels are approximately: 0, 95, 135, 175, 215, 255
	// We'll use thresholds at the midpoints
	if value < 47 { // 0-47 -> level 0
		return 0
	} else if value < 115 { // 48-114 -> level 1
		return 1
	} else if value < 155 { // 115-154 -> level 2
		return 2
	} else if value < 195 { // 155-194 -> level 3
		return 3
	} else if value < 235 { // 195-234 -> level 4
		return 4
	} else { // 235-255 -> level 5
		return 5
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// detectBackgroundColor analyzes edge pixels to identify the most common background color.
// This helps distinguish between transparent areas and actual Pokemon colors.
func detectBackgroundColor(img image.Image) imgcolor.Color {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Sample edge pixels to find the most common color
	colorCounts := make(map[uint32]int)
	var samples []imgcolor.Color

	// Sample top and bottom edges
	for x := 0; x < width; x++ {
		// Top edge
		topColor := img.At(x, 0)
		if r, g, b, a := topColor.RGBA(); a > 0 {
			key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
			colorCounts[key]++
			samples = append(samples, topColor)
		}

		// Bottom edge
		if height > 1 {
			bottomColor := img.At(x, height-1)
			if r, g, b, a := bottomColor.RGBA(); a > 0 {
				key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
				colorCounts[key]++
				samples = append(samples, bottomColor)
			}
		}
	}

	// Sample left and right edges
	for y := 0; y < height; y++ {
		// Left edge
		leftColor := img.At(0, y)
		if r, g, b, a := leftColor.RGBA(); a > 0 {
			key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
			colorCounts[key]++
			samples = append(samples, leftColor)
		}

		// Right edge
		if width > 1 {
			rightColor := img.At(width-1, y)
			if r, g, b, a := rightColor.RGBA(); a > 0 {
				key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
				colorCounts[key]++
				samples = append(samples, rightColor)
			}
		}
	}

	// Find the most common color
	var mostCommonKey uint32
	maxCount := 0
	for key, count := range colorCounts {
		if count > maxCount {
			maxCount = count
			mostCommonKey = key
		}
	}

	// Find a sample of the most common color
	for _, sample := range samples {
		r, g, b, _ := sample.RGBA()
		key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
		if key == mostCommonKey {
			return sample
		}
	}

	// Fallback to a neutral color if no background detected
	return imgcolor.RGBA{240, 240, 240, 255}
}

// isBackgroundColor checks if a pixel matches the detected background color using color distance.
// Uses a threshold to account for slight variations in background color.
func isBackgroundColor(pixel imgcolor.Color, background imgcolor.Color) bool {
	// Get alpha values to check for transparency
	_, _, _, pa := pixel.RGBA()
	_, _, _, ba := background.RGBA()

	// Skip transparent pixels
	if pa < 32768 || ba < 32768 {
		return true
	}

	// Calculate color distance
	distance := colorDistance(pixel, background)

	// Use a threshold for background detection
	// Colors within this distance are considered background
	return distance < 30.0
}

// colorDistance calculates the Euclidean distance between two colors in RGB space.
func colorDistance(c1, c2 imgcolor.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Convert to 8-bit values
	r1_8 := float64(r1 >> 8)
	g1_8 := float64(g1 >> 8)
	b1_8 := float64(b1 >> 8)
	r2_8 := float64(r2 >> 8)
	g2_8 := float64(g2 >> 8)
	b2_8 := float64(b2 >> 8)

	// Calculate Euclidean distance
	dr := r1_8 - r2_8
	dg := g1_8 - g2_8
	db := b1_8 - b2_8

	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// displayPokemon shows the Pokemon info with ASCII art in authentic Pokedex style.
// Mimics the classic Pokedex layout with name/art at top and About/Types sections below.
// Uses type-based colors (Fire=red, Water=blue, Electric=yellow, etc.) for visual appeal.
func displayPokemon(pokemon Pokemon, asciiArt []string) {

	// Display Pokemon name and number (centered above ASCII art)
	nameContent := color.New(color.Bold, color.FgWhite, color.Underline).Sprintf("%s", strings.Title(pokemon.Name))
	numberContent := color.New(color.FgWhite, color.Underline).Sprintf("#%d", pokemon.ID)
	
	// Center the name and number above ASCII art
	nameLineLength := getVisualLength(nameContent + "  " + numberContent)
	namePadding := (asciiWidth - nameLineLength) / 2
	if namePadding < 0 {
		namePadding = 0
	}
	
	fmt.Printf("%s%s  %s\n", strings.Repeat(" ", namePadding), nameContent, numberContent)
	fmt.Println() // Space before ASCII art

	// Display ASCII art centered
	for _, line := range asciiArt {
		fmt.Println(line)
	}

	fmt.Println() // Space after ASCII art

	// Create the bottom info section in Pokedex style
	// Left side: About section, Right side: Types section
	
	// Build About section
	aboutLines := []string{
		color.New(color.Bold, color.Underline).Sprint("About"),
		"",
		fmt.Sprintf("%.1f kg", float64(pokemon.Weight)/10), // Convert hg to kg
		"Weight",
		"",
		fmt.Sprintf("%.1f m", float64(pokemon.Height)/10), // Convert dm to m  
		"Height",
		"",
		fmt.Sprintf("%d", pokemon.BaseExperience),
		"Base Experience",
	}

	// Build Types section
	typesLines := []string{
		color.New(color.Bold, color.Underline).Sprint("Types"),
		"",
	}
	
	// Add types with colors and background
	for _, pokemonType := range pokemon.Types {
		typeName := strings.Title(pokemonType)
		lowerType := strings.ToLower(pokemonType)
		
		// Create type badge with background color and bold text for readability
		var typeBadge string
		switch lowerType {
		case "fire":
			typeBadge = color.New(color.FgWhite, color.BgRed, color.Bold).Sprint(" " + typeName + " ")
		case "water":
			typeBadge = color.New(color.FgWhite, color.BgBlue, color.Bold).Sprint(" " + typeName + " ")
		case "grass":
			typeBadge = color.New(color.FgWhite, color.BgGreen, color.Bold).Sprint(" " + typeName + " ")
		case "electric":
			typeBadge = color.New(color.FgBlack, color.BgYellow, color.Bold).Sprint(" " + typeName + " ")
		case "psychic":
			typeBadge = color.New(color.FgWhite, color.BgMagenta, color.Bold).Sprint(" " + typeName + " ")
		case "ice":
			typeBadge = color.New(color.FgBlack, color.BgCyan, color.Bold).Sprint(" " + typeName + " ")
		case "dragon":
			typeBadge = color.New(color.FgWhite, color.BgMagenta, color.Bold).Sprint(" " + typeName + " ")
		case "dark":
			typeBadge = color.New(color.FgWhite, color.BgBlack, color.Bold).Sprint(" " + typeName + " ")
		case "fighting":
			typeBadge = color.New(color.FgWhite, color.BgRed, color.Bold).Sprint(" " + typeName + " ")
		case "poison":
			typeBadge = color.New(color.FgWhite, color.BgMagenta, color.Bold).Sprint(" " + typeName + " ")
		case "ground":
			typeBadge = color.New(color.FgBlack, color.BgYellow, color.Bold).Sprint(" " + typeName + " ")
		case "flying":
			typeBadge = color.New(color.FgBlack, color.BgCyan, color.Bold).Sprint(" " + typeName + " ")
		case "bug":
			typeBadge = color.New(color.FgWhite, color.BgGreen, color.Bold).Sprint(" " + typeName + " ")
		case "rock":
			typeBadge = color.New(color.FgBlack, color.BgYellow, color.Bold).Sprint(" " + typeName + " ")
		case "ghost":
			typeBadge = color.New(color.FgWhite, color.BgMagenta, color.Bold).Sprint(" " + typeName + " ")
		case "steel":
			typeBadge = color.New(color.FgBlack, color.BgWhite, color.Bold).Sprint(" " + typeName + " ")
		case "fairy":
			typeBadge = color.New(color.FgBlack, color.BgMagenta, color.Bold).Sprint(" " + typeName + " ")
		case "normal":
			typeBadge = color.New(color.FgBlack, color.BgWhite, color.Bold).Sprint(" " + typeName + " ")
		default:
			typeBadge = color.New(color.FgWhite, color.BgBlack, color.Bold).Sprint(" " + typeName + " ")
		}
		
		typesLines = append(typesLines, typeBadge)
	}
	
	typesLines = append(typesLines, "", color.New(color.Bold, color.Underline).Sprint("Stats"))
	
	// Add battle stats
	for _, stat := range pokemon.Stats {
		parts := strings.Split(stat, ": ")
		if len(parts) == 2 {
			statName := strings.Title(parts[0])
			statValue := parts[1]
			
			// Create simple stat display
			statNum := 0
			fmt.Sscanf(statValue, "%d", &statNum)
			barLength := statNum / 20 // Smaller bars for this layout
			if barLength > 10 {
				barLength = 10
			}
			
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", 10-barLength)
			typesLines = append(typesLines, fmt.Sprintf("%s: %s [%s]", statName, statValue, bar))
		}
	}

	// Add abilities if available
	if len(pokemon.Abilities) > 0 {
		typesLines = append(typesLines, "", color.New(color.Bold, color.Underline).Sprint("Abilities"))
		for _, ability := range pokemon.Abilities {
			typesLines = append(typesLines, "• "+strings.Title(ability))
		}
	}

	// Display About and Types sections side by side
	maxLines := len(aboutLines)
	if len(typesLines) > maxLines {
		maxLines = len(typesLines)
	}

	// Ensure both sections have same height
	for len(aboutLines) < maxLines {
		aboutLines = append(aboutLines, "")
	}
	for len(typesLines) < maxLines {
		typesLines = append(typesLines, "")
	}

	// Calculate positioning for centered display (align with ASCII art)
	// TODO: Play with this for centering
	aboutWidth := 20    // About section width
	typesWidth := 35    // Types section width (wider for stats)
	sectionSpacing := 5 // Space between sections (reduced for better centering)
	totalSectionWidth := aboutWidth + sectionSpacing + typesWidth
	sectionPadding := (asciiWidth - totalSectionWidth) / 2
	if sectionPadding < 0 {
		sectionPadding = 0
	}

	// Display both sections side by side
	for i := 0; i < maxLines; i++ {
		aboutLine := aboutLines[i]
		typesLine := typesLines[i]
		
		// Pad about line to consistent width
		if len(aboutLine) > aboutWidth {
			aboutLine = aboutLine[:aboutWidth]
		}
		aboutPadding := aboutWidth - getVisualLength(aboutLine)
		
		fmt.Printf("%s%-*s%s%s\n", 
			strings.Repeat(" ", sectionPadding),
			aboutWidth, aboutLine+strings.Repeat(" ", aboutPadding),
			strings.Repeat(" ", sectionSpacing),
			typesLine)
	}
}

// displayPokemonTextOnly shows Pokemon info without ASCII art for narrow terminals
func displayPokemonTextOnly(pokemon Pokemon) {
	// Colors for different types
	typeColors := map[string]*color.Color{
		"fire":     color.New(color.FgRed),
		"water":    color.New(color.FgBlue),
		"grass":    color.New(color.FgGreen),
		"electric": color.New(color.FgYellow),
		"psychic":  color.New(color.FgMagenta),
		"ice":      color.New(color.FgCyan),
		"dragon":   color.New(color.FgMagenta),
		"dark":     color.New(color.FgBlack),
		"fighting": color.New(color.FgRed, color.Bold),
		"poison":   color.New(color.FgMagenta),
		"ground":   color.New(color.FgYellow, color.Bold),
		"flying":   color.New(color.FgCyan),
		"bug":      color.New(color.FgGreen),
		"rock":     color.New(color.FgYellow, color.Bold),
		"ghost":    color.New(color.FgMagenta),
		"steel":    color.New(color.FgWhite, color.Bold),
		"fairy":    color.New(color.FgMagenta, color.Bold),
		"normal":   color.New(color.FgWhite),
	}

	// Simple text-only display for narrow terminals
	fmt.Printf("\n%s\n", color.New(color.Bold).Sprintf("=== %s (#%d) ===", strings.Title(pokemon.Name), pokemon.ID))

	fmt.Printf("Height: %d dm\n", pokemon.Height)
	fmt.Printf("Weight: %d hg\n", pokemon.Weight)
	fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)

	// Display types with colors
	if len(pokemon.Types) > 0 {
		typeStr := "Type: "
		for i, pokemonType := range pokemon.Types {
			if i > 0 {
				typeStr += ", "
			}
			if typeColor, exists := typeColors[strings.ToLower(pokemonType)]; exists {
				typeStr += typeColor.Sprint(strings.Title(pokemonType))
			} else {
				typeStr += strings.Title(pokemonType)
			}
		}
		fmt.Println(typeStr)
	}

	// Display abilities
	if len(pokemon.Abilities) > 0 {
		fmt.Printf("Abilities: %s\n", strings.Join(pokemon.Abilities, ", "))
	}

	// Display stats with simple bars
	if len(pokemon.Stats) > 0 {
		fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("STATS:"))
		for _, stat := range pokemon.Stats {
			parts := strings.Split(stat, ": ")
			if len(parts) == 2 {
				statName := parts[0]
				statValue := parts[1]

				// Create a simple stat bar for narrow terminal
				statNum := 0
				fmt.Sscanf(statValue, "%d", &statNum)
				barLength := statNum / 20 // Smaller scale for narrow terminal
				if barLength > 10 {
					barLength = 10
				}

				bar := strings.Repeat("█", barLength) + strings.Repeat("░", 10-barLength)
				fmt.Printf("%s: %s [%s]\n", statName, statValue, bar)
			}
		}
	}

	fmt.Println() // Extra spacing
}
