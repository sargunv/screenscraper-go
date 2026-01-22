package chd

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ulikunitz/xz/lzma"
)

// CD-ROM frame constants (from libchdr/cdrom.h)
const (
	CDMaxSectorData  = 2352 // Raw sector data size
	CDMaxSubcodeData = 96   // Subcode data per frame
	CDFrameSize      = 2448 // CDMaxSectorData + CDMaxSubcodeData
	CDSyncSize       = 12   // Sync pattern size
	CDDataOffset     = 16   // User data offset in Mode 1
	CDUserDataSize   = 2048 // User data size in Mode 1/2
)

// CD-ROM sync pattern
var cdSyncPattern = []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}

// decompressCDZlib decompresses CD-ROM data with zlib base compression.
func decompressCDZlib(data []byte, hunkBytes uint32) ([]byte, error) {
	return decompressCDCodec(data, hunkBytes, decompressZlibRaw, "zlib")
}

// decompressCDLZMA decompresses CD-ROM data with LZMA base compression.
func decompressCDLZMA(data []byte, hunkBytes uint32) ([]byte, error) {
	return decompressCDCodec(data, hunkBytes, decompressLZMARaw, "lzma")
}

// decompressCDZstd decompresses CD-ROM data with Zstandard base compression.
func decompressCDZstd(data []byte, hunkBytes uint32) ([]byte, error) {
	return decompressCDCodec(data, hunkBytes, decompressZstdRaw, "zstd")
}

// decompressCDCodec is the common implementation for CD codecs.
// Format: [ECC bitmap] [compressed base length] [base data (sector)] [subcode data (zlib)]
func decompressCDCodec(data []byte, hunkBytes uint32, baseDecompress func([]byte, int) ([]byte, error), codecName string) ([]byte, error) {
	// Calculate frame count
	frames := int(hunkBytes / CDFrameSize)
	if frames == 0 {
		return nil, fmt.Errorf("CD codec: invalid hunk size %d", hunkBytes)
	}

	// Header format:
	// - ECC bitmap: (frames + 7) / 8 bytes
	// - Compressed base length: 2 bytes if hunkBytes < 65536, else 3 bytes
	eccBytes := (frames + 7) / 8
	complenBytes := 2
	if hunkBytes >= 65536 {
		complenBytes = 3
	}
	headerBytes := eccBytes + complenBytes

	if len(data) < headerBytes {
		return nil, fmt.Errorf("CD codec: data too short for header (need %d, have %d)", headerBytes, len(data))
	}

	// Extract compressed base length
	var complenBase int
	if complenBytes == 2 {
		complenBase = int(data[eccBytes])<<8 | int(data[eccBytes+1])
	} else {
		complenBase = int(data[eccBytes])<<16 | int(data[eccBytes+1])<<8 | int(data[eccBytes+2])
	}

	if len(data) < headerBytes+complenBase {
		return nil, fmt.Errorf("CD codec: data too short for base (need %d, have %d)", headerBytes+complenBase, len(data))
	}

	// Decompress base (sector) data - outputs CDMaxSectorData (2352) bytes per frame
	baseCompressed := data[headerBytes : headerBytes+complenBase]
	expectedBaseSize := frames * CDMaxSectorData
	baseData, err := baseDecompress(baseCompressed, expectedBaseSize)
	if err != nil {
		return nil, fmt.Errorf("CD codec base decompress (%s): %w", codecName, err)
	}

	// Decompress subcode data (always zlib) - outputs CDMaxSubcodeData (96) bytes per frame
	subcodeCompressed := data[headerBytes+complenBase:]
	expectedSubcodeSize := frames * CDMaxSubcodeData
	var subcodeData []byte
	if len(subcodeCompressed) > 0 {
		subcodeData, err = decompressZlibRaw(subcodeCompressed, expectedSubcodeSize)
		if err != nil {
			return nil, fmt.Errorf("CD codec subcode decompress: %w", err)
		}
	} else {
		// No subcode data - fill with zeros
		subcodeData = make([]byte, expectedSubcodeSize)
	}

	// Reconstruct frames: interleave sector data and subcode data
	result := make([]byte, hunkBytes)
	for i := range frames {
		// Copy sector data (2352 bytes)
		srcOffset := i * CDMaxSectorData
		dstOffset := i * CDFrameSize
		if srcOffset+CDMaxSectorData <= len(baseData) {
			copy(result[dstOffset:], baseData[srcOffset:srcOffset+CDMaxSectorData])
		}

		// Copy subcode data (96 bytes)
		srcSubOffset := i * CDMaxSubcodeData
		dstSubOffset := dstOffset + CDMaxSectorData
		if srcSubOffset+CDMaxSubcodeData <= len(subcodeData) {
			copy(result[dstSubOffset:], subcodeData[srcSubOffset:srcSubOffset+CDMaxSubcodeData])
		}
	}

	return result, nil
}

// Raw decompression functions (without CHD-specific headers)

func decompressZlibRaw(data []byte, outputSize int) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return result[:n], nil
}

func decompressLZMARaw(data []byte, outputSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("LZMA data empty")
	}

	// CHD CD LZMA uses default properties: lc=3, lp=0, pb=2
	// Properties byte = (pb * 5 + lp) * 9 + lc = (2*5 + 0)*9 + 3 = 93 = 0x5D
	propsByte := byte(0x5D)

	// Dictionary size: use a reasonable size for CD data
	dictSize := uint32(65536)
	if outputSize > 65536 {
		dictSize = uint32(outputSize)
	}

	// Build standard LZMA header
	header := make([]byte, 13)
	header[0] = propsByte
	binary.LittleEndian.PutUint32(header[1:5], dictSize)
	binary.LittleEndian.PutUint64(header[5:13], uint64(outputSize))

	fullData := append(header, data...)

	r, err := lzma.NewReader(bytes.NewReader(fullData))
	if err != nil {
		return nil, err
	}

	result := make([]byte, outputSize)
	n, err := io.ReadFull(r, result)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return result[:n], nil
}

func decompressZstdRaw(data []byte, outputSize int) ([]byte, error) {
	result, err := zstdDecoder.DecodeAll(data, make([]byte, 0, outputSize))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExtractUserData extracts the 2048-byte user data portion from a CD-ROM sector.
// For Mode 1 sectors, this is bytes 16-2063 of the 2352-byte frame.
func ExtractUserData(sector []byte) []byte {
	if len(sector) < CDMaxSectorData {
		return sector // Return as-is for non-CD data
	}

	// Check if this looks like a CD-ROM sector with sync pattern
	if len(sector) >= CDSyncSize {
		isSync := true
		for i, b := range cdSyncPattern {
			if sector[i] != b {
				isSync = false
				break
			}
		}
		if isSync {
			// Extract user data from Mode 1 sector (bytes 16-2063)
			if len(sector) >= CDDataOffset+CDUserDataSize {
				return sector[CDDataOffset : CDDataOffset+CDUserDataSize]
			}
		}
	}

	// For non-Mode1 or unrecognized format, return first 2048 bytes
	if len(sector) >= CDUserDataSize {
		return sector[:CDUserDataSize]
	}
	return sector
}
