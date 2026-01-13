// Package romident provides ROM identification and hashing utilities.
package romident

import "github.com/sargunv/rom-tools/lib/romident/game"

// ROMType indicates how the ROM is packaged.
type ROMType string

const (
	ROMTypeFile   ROMType = "file"
	ROMTypeZIP    ROMType = "zip"
	ROMTypeFolder ROMType = "folder"
)

// Format is an alias for game.Format.
type Format = game.Format

// Format constants re-exported from game package.
const (
	FormatUnknown = game.FormatUnknown
	FormatCHD     = game.FormatCHD
	FormatXISO    = game.FormatXISO
	FormatXBE     = game.FormatXBE
	FormatISO9660 = game.FormatISO9660
	FormatZIP     = game.FormatZIP
	FormatGBA     = game.FormatGBA
	FormatZ64     = game.FormatZ64
	FormatV64     = game.FormatV64
	FormatN64     = game.FormatN64
	FormatGB      = game.FormatGB
	FormatMD      = game.FormatMD
	FormatSMD     = game.FormatSMD
	FormatNDS     = game.FormatNDS
	FormatNES     = game.FormatNES
	FormatSNES    = game.FormatSNES
	FormatGCM     = game.FormatGCM
	FormatRVZ     = game.FormatRVZ
)

// HashAlgorithm identifies a hash algorithm.
type HashAlgorithm string

const (
	HashSHA1  HashAlgorithm = "sha1"
	HashMD5   HashAlgorithm = "md5"
	HashCRC32 HashAlgorithm = "crc32"
)

// HashSource indicates where a hash value came from.
type HashSource string

const (
	HashSourceCalculated    HashSource = "calculated"
	HashSourceZIPMetadata   HashSource = "zip-metadata"
	HashSourceCHDRaw        HashSource = "chd-raw"
	HashSourceCHDCompressed HashSource = "chd-compressed"
)

// Hash represents a computed or extracted hash value.
type Hash struct {
	Algorithm HashAlgorithm `json:"algorithm"`
	Value     string        `json:"value"` // hex-encoded
	Source    HashSource    `json:"source"`
}

// ROMFile represents a single file within a ROM.
type ROMFile struct {
	Size      int64  `json:"size"`
	Format    Format `json:"format"`
	Hashes    []Hash `json:"hashes"`
	IsPrimary bool   `json:"is_primary,omitempty"` // true if used for identification
}

type GameIdent = game.GameIdent
type Region = game.Region
type Platform = game.Platform

// Files is a map of file path to file info.
type Files map[string]ROMFile

// ROM represents a complete game unit (single file, zip, or folder).
type ROM struct {
	Path  string     `json:"path"`
	Type  ROMType    `json:"type"`
	Files Files      `json:"files"`
	Ident *GameIdent `json:"ident,omitempty"`
}

// HashMode controls how hashes are calculated.
type HashMode int

const (
	// HashModeDefault uses fast methods where available (CHD header, ZIP metadata),
	// calculates full hashes for loose files.
	HashModeDefault HashMode = iota

	// HashModeFast skips hash calculation for large files, but still calculates
	// hashes for small files (below FastModeSmallFileThreshold).
	HashModeFast

	// HashModeSlow calculates full hashes even when fast methods are available
	// (e.g., decompresses ZIP files to calculate SHA1/MD5).
	HashModeSlow
)

const (
	// FastModeSmallFileThreshold is the size threshold below which fast mode
	// will still calculate hashes. Files at or above this size skip hash calculation.
	// 65 MiB covers most cartridge ROMs (GBA, SNES, NES, etc.) but skips large disc images.
	FastModeSmallFileThreshold = 65 * 1024 * 1024 // 65 MiB
)

// Options controls ROM identification behavior.
type Options struct {
	HashMode HashMode
}
