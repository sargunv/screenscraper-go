package scraper

import (
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/screenscraper"
)

// BaseName returns the filename without extension
func BaseName(filename string) string {
	if ext := filepath.Ext(filename); ext != "" {
		return strings.TrimSuffix(filename, ext)
	}
	return filename
}

// Hashes contains ROM hash values
type Hashes struct {
	SHA1  string
	MD5   string
	CRC32 string
}

// CacheKey returns the primary key for cache lookups
func (h Hashes) CacheKey() string {
	if h.SHA1 != "" {
		return "sha1:" + strings.ToLower(h.SHA1)
	}
	if h.MD5 != "" {
		return "md5:" + strings.ToLower(h.MD5)
	}
	return "crc32:" + strings.ToLower(h.CRC32)
}

// IsEmpty returns true if no hashes are set
func (h Hashes) IsEmpty() bool {
	return h.SHA1 == "" && h.MD5 == "" && h.CRC32 == ""
}

// LookupSource indicates where the lookup entry came from
type LookupSource int

const (
	SourceDAT LookupSource = iota
	SourceROM
)

// LookupEntry is the unified input for Screenscraper lookups
type LookupEntry struct {
	// Identification
	Name     string // Display name (from DAT or filename)
	FileName string // ROM filename with extension (for API)
	Hashes   Hashes // SHA1, MD5, CRC32
	Serial   string // Game code (from DAT serial or ROM header)
	Size     int64  // File size in bytes

	// Region info (parsed from name or ROM header)
	Regions []string // e.g., ["us", "eu"]

	// Output path (for media naming)
	BaseName string // Filename without extension

	// Source info
	Source  LookupSource
	ROMPath string // Only for ROM source
}

// ScrapeResult contains the result of looking up a single entry
type ScrapeResult struct {
	Entry     *LookupEntry
	Game      *screenscraper.Game // nil if not found
	Media     map[string]string   // mediaType -> local path (downloaded)
	Error     error
	Cached    bool   // true if game info from cache
	Skipped   bool   // true if skipped (BIOS, etc.)
	Reason    string // reason for skip/error
	CacheHits int    // total API calls avoided due to cache
}

// Status represents the status of a scrape operation
type Status int

const (
	StatusPending Status = iota
	StatusInProgress
	StatusFound
	StatusNotFound
	StatusSkipped
	StatusError
)

// String returns a string representation of the status
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusInProgress:
		return "in_progress"
	case StatusFound:
		return "found"
	case StatusNotFound:
		return "not_found"
	case StatusSkipped:
		return "skipped"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Config holds scraper configuration
type Config struct {
	// System
	SystemID string

	// Media selection
	MediaTypes []string // ES-DE media type names

	// Region preferences
	PreferredRegions []string

	// Output directory for media files
	MediaOutputDir string

	// Cache settings
	SkipCacheRead  bool // --no-cache
	SkipCacheWrite bool // --cache-only

	// Overwrite behavior
	Overwrite bool // --overwrite

	// Rate limiting (from user info)
	MaxThreads        int
	MaxRequestsPerMin int

	// Filter for which entries to scrape
	Filter       *Filter
	FilterConfig *FilterConfig
}

// DefaultMediaTypes returns the default media types to download
func DefaultMediaTypes() []string {
	return []string{"screenshots", "covers", "marquees"}
}

// AllMediaTypes returns all supported media types
func AllMediaTypes() []string {
	return []string{
		"screenshots",
		"titlescreens",
		"covers",
		"3dboxes",
		"marquees",
		"fanart",
		"videos",
		"physicalmedia",
		"backcovers",
	}
}

// MediaTypeMapping maps ES-DE folder names to Screenscraper media types
// Some have fallbacks (e.g., marquees tries wheel-hd first, then wheel)
var MediaTypeMapping = map[string][]string{
	"screenshots":   {"ss"},
	"titlescreens":  {"sstitle"},
	"covers":        {"box-2D"},
	"3dboxes":       {"box-3D"},
	"marquees":      {"wheel-hd", "wheel"},
	"fanart":        {"fanart"},
	"videos":        {"video-normalized", "video"},
	"physicalmedia": {"support-2D"},
	"backcovers":    {"box-2D-back"},
}

// MediaExtensions maps media types to expected file extensions
var MediaExtensions = map[string]string{
	"ss":               "png",
	"sstitle":          "png",
	"box-2D":           "png",
	"box-3D":           "png",
	"wheel-hd":         "png",
	"wheel":            "png",
	"fanart":           "jpg",
	"video-normalized": "mp4",
	"video":            "mp4",
	"support-2D":       "png",
	"box-2D-back":      "png",
}
