// Package sprites provides simple sprite downloading and caching functionality.
// This package implements a "just works" approach - no configuration needed.
// Sprites are automatically cached in ~/.pokedex_sprites/ for instant re-display.
package sprites

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const spriteCacheDir = ".pokedex_sprites"

// DownloadAndCacheSprite downloads Pokemon sprites and caches them locally for instant re-display.
//
// This function implements a simple but effective caching strategy:
// - Uses MD5 hash of URL as cache filename to avoid duplicates
// - Caches in user's home directory under ~/.pokedex_sprites/
// - Returns cached version if available, otherwise downloads fresh copy
// - Gracefully handles errors by continuing without caching if home dir unavailable
//
// Returns the sprite data as []byte or an error if download fails.
func DownloadAndCacheSprite(url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("empty sprite URL")
	}

	// Create cache directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, just download without caching
		return downloadSprite(url)
	}

	cacheDir := filepath.Join(homeDir, spriteCacheDir)
	os.MkdirAll(cacheDir, 0755) // Create if doesn't exist, ignore errors

	// Generate simple cache filename from URL hash
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	cacheFile := filepath.Join(cacheDir, hash+".png")

	// Check if already cached
	if data, err := os.ReadFile(cacheFile); err == nil {
		return data, nil
	}

	// Download the sprite
	data, err := downloadSprite(url)
	if err != nil {
		return nil, err
	}

	// Try to cache it (don't fail if caching fails)
	os.WriteFile(cacheFile, data, 0644)

	return data, nil
}

// downloadSprite performs a simple HTTP download with a reasonable 10-second timeout.
// This prevents the application from hanging on slow network connections while
// still allowing enough time for sprite downloads to complete.
func downloadSprite(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download sprite: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sprite: %w", err)
	}

	return data, nil
}
