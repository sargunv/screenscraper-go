// Package identify provides ROM identification and hashing utilities.
package identify

import "github.com/sargunv/rom-tools/lib/core"

// HashType identifies a specific hash (combines algorithm and source).
type HashType string

const (
	HashSHA1                HashType = "sha1"
	HashMD5                 HashType = "md5"
	HashCRC32               HashType = "crc32"
	HashZipCRC32            HashType = "zip-crc32"
	HashCHDUncompressedSHA1 HashType = "chd-uncompressed-sha1"
	HashCHDCompressedSHA1   HashType = "chd-compressed-sha1"
)

// Hashes maps hash type to hex-encoded value.
type Hashes map[HashType]string

// Item represents one identifiable unit (a file or entry within a container).
type Item struct {
	Name   string        `json:"name"`             // filename (basename for single files, relative path in containers)
	Size   int64         `json:"size"`             // file size in bytes
	Hashes Hashes        `json:"hashes,omitempty"` // hash values by type
	Game   core.GameInfo `json:"game,omitempty"`   // identified game info (platform-specific struct)
}

// Result is the result of identifying a path.
type Result struct {
	Path  string `json:"path"`  // absolute path that was identified
	Items []Item `json:"items"` // identified items (1 for single file, N for containers)
}

// Options controls ROM identification behavior.
type Options struct {
	// MaxHashSize is the maximum file size (in bytes) for which hashes will be calculated.
	// Files larger than this skip hash calculation entirely.
	// Use -1 for no limit (always calculate hashes).
	// Default is 64 MiB.
	MaxHashSize int64

	// DecompressArchives controls whether compressed archives (e.g., ZIP) are
	// decompressed to calculate hashes and identify games inside.
	// If false, only archive metadata (pre-computed CRC32, file sizes) is used.
	// Default is true.
	DecompressArchives bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxHashSize:        64 * 1024 * 1024, // 64 MiB
		DecompressArchives: true,
	}
}
