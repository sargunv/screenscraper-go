package rvz

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/gcm"
)

// RVZ/WIA (Dolphin compressed disc image) format parsing.
//
// Format specification:
// https://raw.githubusercontent.com/dolphin-emu/dolphin/refs/heads/master/docs/WiaAndRvz.md
//
// wia_file_head_t (offset 0x0, 0x48 bytes):
//
//	Offset  Size  Description
//	0x00    4     Magic: "WIA\x1"
//	0x04    4     Version
//	0x08    4     Compatible version
//	0x0C    4     Disc struct size
//	0x10    20    SHA-1 hash of disc struct
//	0x24    8     ISO file size (original disc size)
//	0x2C    8     WIA/RVZ file size
//	0x34    20    SHA-1 hash of file header
//
// wia_disc_t (offset 0x48):
//
//	Offset  Size  Description
//	0x00    4     Disc type (1=GameCube, 2=Wii)
//	0x04    4     Compression method (0=NONE, 2=BZIP2, 3=LZMA, 4=LZMA2, 5=Zstandard)
//	0x08    4     Compression level (signed for Zstandard)
//	0x0C    4     Chunk size
//	0x10    128   dhead[0x80] - First 128 bytes of disc (UNCOMPRESSED!)

const (
	fileHeadSize   = 0x48
	discStructBase = 0x48

	// wia_file_head_t offsets
	magicOffset        = 0x00
	versionOffset      = 0x04
	compatVerOffset    = 0x08
	discSizeOffset     = 0x0C
	discHashOffset     = 0x10
	isoFileSizeOffset  = 0x24
	wiaFileSizeOffset  = 0x2C
	fileHeadHashOffset = 0x34

	// wia_disc_t offsets (relative to discStructBase)
	discTypeOffset    = 0x00
	compressionOffset = 0x04
	comprLevelOffset  = 0x08
	chunkSizeOffset   = 0x0C
	dheadOffset       = 0x10
	dheadSize         = 0x80

	// Total header size we need to read
	totalHeaderSize = discStructBase + dheadOffset + dheadSize
)

// Compression method constants
const (
	CompressionNone      = 0
	CompressionPurge     = 1 // Not used in RVZ
	CompressionBZIP2     = 2
	CompressionLZMA      = 3
	CompressionLZMA2     = 4
	CompressionZstandard = 5
)

// Disc type constants
const (
	DiscTypeUnknown  = 0
	DiscTypeGameCube = 1
	DiscTypeWii      = 2
)

// RVZInfo contains metadata extracted from an RVZ/WIA file header.
type RVZInfo struct {
	Version           uint32 // WIA format version
	CompatibleVersion uint32 // Minimum compatible version
	DiscType          int    // 1=GameCube, 2=Wii
	Compression       string // Compression method name
	CompressionLevel  int32  // Compression level (signed for Zstandard)
	ChunkSize         uint32 // Chunk size for compressed data
	ISOFileSize       uint64 // Original uncompressed disc size
	WIAFileSize       uint64 // Compressed file size
	DiscHash          string // SHA-1 hash of disc struct (hex)
	FileHeadHash      string // SHA-1 hash of file header (hex)
	DiscHeader        []byte // First 128 bytes of disc (uncompressed)
}

// RVZExtraInfo combines GCInfo and RVZInfo for the Extra field.
type RVZExtraInfo struct {
	GCInfo  *gcm.GCInfo
	RVZInfo *RVZInfo
}

// ParseRVZHeader reads and parses an RVZ/WIA file header.
// Returns the RVZ metadata including the uncompressed disc header bytes.
func ParseRVZHeader(r io.ReaderAt, size int64) (*RVZInfo, error) {
	if size < totalHeaderSize {
		return nil, fmt.Errorf("file too small for RVZ header: need %d bytes, got %d", totalHeaderSize, size)
	}

	header := make([]byte, totalHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read RVZ header: %w", err)
	}

	// Verify magic bytes "WIA\x1" or "RVZ\x1"
	magic := string(header[magicOffset : magicOffset+4])
	if magic != "WIA\x01" && magic != "RVZ\x01" {
		return nil, fmt.Errorf("not a valid RVZ/WIA file: invalid magic (got %q)", magic)
	}

	// Parse wia_file_head_t
	version := binary.BigEndian.Uint32(header[versionOffset:])
	compatVer := binary.BigEndian.Uint32(header[compatVerOffset:])
	discHash := hex.EncodeToString(header[discHashOffset : discHashOffset+20])
	isoFileSize := binary.BigEndian.Uint64(header[isoFileSizeOffset:])
	wiaFileSize := binary.BigEndian.Uint64(header[wiaFileSizeOffset:])
	fileHeadHash := hex.EncodeToString(header[fileHeadHashOffset : fileHeadHashOffset+20])

	// Parse wia_disc_t
	discType := int(binary.BigEndian.Uint32(header[discStructBase+discTypeOffset:]))
	compression := binary.BigEndian.Uint32(header[discStructBase+compressionOffset:])
	comprLevel := int32(binary.BigEndian.Uint32(header[discStructBase+comprLevelOffset:]))
	chunkSize := binary.BigEndian.Uint32(header[discStructBase+chunkSizeOffset:])

	// Extract dhead (first 128 bytes of disc, uncompressed)
	dhead := make([]byte, dheadSize)
	copy(dhead, header[discStructBase+dheadOffset:])

	return &RVZInfo{
		Version:           version,
		CompatibleVersion: compatVer,
		DiscType:          discType,
		Compression:       compressionName(compression),
		CompressionLevel:  comprLevel,
		ChunkSize:         chunkSize,
		ISOFileSize:       isoFileSize,
		WIAFileSize:       wiaFileSize,
		DiscHash:          discHash,
		FileHeadHash:      fileHeadHash,
		DiscHeader:        dhead,
	}, nil
}

// Identify verifies the format and extracts game identification from an RVZ/WIA file.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	rvzInfo, err := ParseRVZHeader(r, size)
	if err != nil {
		return nil, err
	}

	gcInfo, err := gcm.ParseDiscHeader(rvzInfo.DiscHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse disc header from RVZ: %w", err)
	}

	extra := &RVZExtraInfo{
		GCInfo:  gcInfo,
		RVZInfo: rvzInfo,
	}

	return gcm.GCInfoToGameIdent(gcInfo, extra), nil
}

// compressionName returns a human-readable name for the compression method.
func compressionName(method uint32) string {
	switch method {
	case CompressionNone:
		return "none"
	case CompressionPurge:
		return "purge"
	case CompressionBZIP2:
		return "bzip2"
	case CompressionLZMA:
		return "lzma"
	case CompressionLZMA2:
		return "lzma2"
	case CompressionZstandard:
		return "zstd"
	default:
		return fmt.Sprintf("unknown(%d)", method)
	}
}
