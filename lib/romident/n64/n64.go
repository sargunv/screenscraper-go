package n64

import (
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/z64"
)

// N64 (word-swapped/little-endian) ROM format.
//
// N64 files have 32-bit words reversed compared to native Z64 format.
// This package converts N64 to Z64 format and delegates to the z64 package.

// Identify verifies the format and extracts game identification from an N64 ROM.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	if size < z64.N64HeaderSize {
		return nil, fmt.Errorf("file too small for N64 header: %d bytes", size)
	}

	// Read first 4 bytes to check byte order
	first4 := make([]byte, 4)
	if _, err := r.ReadAt(first4, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	actualOrder := z64.DetectByteOrder(first4)
	if actualOrder != z64.N64LittleEndian {
		return nil, fmt.Errorf("byte order mismatch: expected n64, got %s", actualOrder)
	}

	// Read full header
	header := make([]byte, z64.N64HeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read N64 header: %w", err)
	}

	// Convert to big-endian (z64) format
	z64.SwapBytes32(header)

	info, err := z64.ParseN64Header(header, z64.N64LittleEndian)
	if err != nil {
		return nil, err
	}

	return z64.N64InfoToGameIdent(info), nil
}
