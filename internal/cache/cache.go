package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Mode determines cache read/write behavior
type Mode int

const (
	// ModeNormal reads from and writes to the cache
	ModeNormal Mode = iota
	// ModeNoRead skips cache reads but still writes (--no-cache)
	ModeNoRead
	// ModeReadOnly reads from cache but doesn't write (--cache-only)
	ModeReadOnly
)

// DiskCache provides file-based caching for game info and media
type DiskCache struct {
	baseDir string
	maxAge  time.Duration
	mode    Mode
}

// New creates a new disk cache
func New(baseDir string, maxAge time.Duration, mode Mode) (*DiskCache, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &DiskCache{
		baseDir: baseDir,
		maxAge:  maxAge,
		mode:    mode,
	}, nil
}

// DefaultCacheDir returns the default cache directory path
func DefaultCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache directory: %w", err)
	}
	return filepath.Join(cacheDir, "rom-tools", "screenscraper", "v1"), nil
}

// hashKey creates a safe filename from a cache key
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:16]) // Use first 16 bytes (32 hex chars)
}

// GameInfoKey creates a cache key for game info
func GameInfoKey(systemID, hash string) string {
	return fmt.Sprintf("game:%s:%s", systemID, strings.ToLower(hash))
}

// MediaKey creates a cache key for media
func MediaKey(systemID, gameID, mediaType, region string) string {
	return fmt.Sprintf("media:%s:%s:%s:%s", systemID, gameID, mediaType, region)
}

// cacheEntry represents metadata about a cached item
type cacheEntry struct {
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

// GetGameInfo retrieves cached game info
func (c *DiskCache) GetGameInfo(systemID, hash string) ([]byte, bool) {
	if c.mode == ModeNoRead {
		return nil, false
	}

	key := GameInfoKey(systemID, hash)
	path := filepath.Join(c.baseDir, "games", systemID, hashKey(key)+".json")

	data, err := c.readIfValid(path)
	if err != nil {
		return nil, false
	}

	return data, true
}

// SetGameInfo stores game info in the cache
func (c *DiskCache) SetGameInfo(systemID, hash string, data []byte) error {
	if c.mode == ModeReadOnly {
		return nil
	}

	key := GameInfoKey(systemID, hash)
	dir := filepath.Join(c.baseDir, "games", systemID)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	path := filepath.Join(dir, hashKey(key)+".json")
	return c.writeWithMeta(path, key, data)
}

// GetMedia retrieves cached media file
// Returns the data and file extension if found
func (c *DiskCache) GetMedia(systemID, gameID, mediaType, region string) ([]byte, string, bool) {
	if c.mode == ModeNoRead {
		return nil, "", false
	}

	key := MediaKey(systemID, gameID, mediaType, region)
	dir := filepath.Join(c.baseDir, "media", systemID, gameID)
	baseName := hashKey(key)

	// Try common extensions (including .nomedia for cached "not available")
	for _, ext := range []string{".nomedia", ".png", ".jpg", ".mp4"} {
		path := filepath.Join(dir, baseName+ext)
		data, err := c.readIfValid(path)
		if err == nil {
			return data, ext[1:], true // Strip the leading dot
		}
	}

	return nil, "", false
}

// SetMedia stores media in the cache
func (c *DiskCache) SetMedia(systemID, gameID, mediaType, region string, data []byte, ext string) error {
	if c.mode == ModeReadOnly {
		return nil
	}

	key := MediaKey(systemID, gameID, mediaType, region)
	dir := filepath.Join(c.baseDir, "media", systemID, gameID)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	path := filepath.Join(dir, hashKey(key)+"."+ext)
	return c.writeWithMeta(path, key, data)
}

// readIfValid reads a file if it exists and hasn't expired
func (c *DiskCache) readIfValid(path string) ([]byte, error) {
	metaPath := path + ".meta"

	// Check metadata for expiration
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(metaData, &entry); err != nil {
		return nil, err
	}

	// Check if expired
	if time.Since(entry.CreatedAt) > c.maxAge {
		// Remove expired files
		os.Remove(path)
		os.Remove(metaPath)
		return nil, fmt.Errorf("cache entry expired")
	}

	// Read the actual data
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// writeWithMeta writes data and metadata files
func (c *DiskCache) writeWithMeta(path, key string, data []byte) error {
	// Write data file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Write metadata
	entry := cacheEntry{
		Key:       key,
		CreatedAt: time.Now(),
	}
	metaData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache metadata: %w", err)
	}

	metaPath := path + ".meta"
	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write cache metadata: %w", err)
	}

	return nil
}

// Clear removes all cached data
func (c *DiskCache) Clear() error {
	return os.RemoveAll(c.baseDir)
}
