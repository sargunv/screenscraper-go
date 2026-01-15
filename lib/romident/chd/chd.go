package chd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// CHD (Compressed Hunks of Data) is MAME's compressed disc image format.
// Format specification: https://github.com/mamedev/mame/blob/master/src/lib/util/chd.h
//
// V5 header layout (124 bytes):
//
//	Offset  Size  Description
//	0       8     Magic ("MComprHD")
//	8       4     Header length (big-endian)
//	12      4     Version (big-endian)
//	16      4     Compressors[0]
//	20      4     Compressors[1]
//	24      4     Compressors[2]
//	28      4     Compressors[3]
//	32      8     Logical bytes (big-endian)
//	40      8     Map offset (big-endian)
//	48      8     Metadata offset (big-endian)
//	56      4     Hunk bytes (big-endian)
//	60      4     Unit bytes (big-endian)
//	64      20    Raw SHA1 (of the raw, uncompressed data)
//	84      20    SHA1 (of the compressed data)
//	104     20    Parent SHA1 (all zeros if no parent)
const (
	headerSize       = 124
	rawSHA1Offset    = 64
	sha1Offset       = 84
	parentSHA1Offset = 104
	sha1Size         = 20
)

// Codec IDs (4-byte ASCII codes stored as big-endian uint32)
const (
	CodecNone   uint32 = 0
	CodecZlib   uint32 = 0x7a6c6962 // 'zlib'
	CodecLZMA   uint32 = 0x6c7a6d61 // 'lzma'
	CodecHuff   uint32 = 0x68756666 // 'huff'
	CodecFLAC   uint32 = 0x666c6163 // 'flac'
	CodecZstd   uint32 = 0x7a737464 // 'zstd'
	CodecCDZlib uint32 = 0x63647a6c // 'cdzl'
	CodecCDLZMA uint32 = 0x63646c7a // 'cdlz'
	CodecCDFLAC uint32 = 0x6364666c // 'cdfl'
	CodecCDZstd uint32 = 0x63647a73 // 'cdzs'
)

// Header contains metadata extracted from a CHD file header.
type Header struct {
	Version      uint32
	Compressors  [4]uint32 // Up to 4 compression codecs
	LogicalBytes uint64    // Total uncompressed size
	MapOffset    uint64    // Offset to hunk map
	MetaOffset   uint64    // Offset to metadata
	HunkBytes    uint32    // Bytes per hunk
	UnitBytes    uint32    // Bytes per unit (sector size)
	TotalHunks   uint32    // Calculated: LogicalBytes / HunkBytes
	RawSHA1      string    // SHA1 of the raw (uncompressed) data
	SHA1         string    // SHA1 of the compressed data
	ParentSHA1   string    // SHA1 of parent CHD (for diffs), empty if standalone
}

// CHDInfo is an alias for Header for backward compatibility.
type CHDInfo = Header

// ParseCHDHeader reads and parses a CHD file header.
// Returns both the raw and compressed SHA1 hashes stored in the header.
func ParseCHDHeader(r io.ReaderAt) (*Header, error) {
	header := make([]byte, headerSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read CHD header: %w", err)
	}

	// Verify magic
	if string(header[0:8]) != "MComprHD" {
		return nil, fmt.Errorf("not a valid CHD file: invalid magic")
	}

	// Header length (big-endian uint32 at offset 8)
	headerLen := binary.BigEndian.Uint32(header[8:12])

	// Version (big-endian uint32 at offset 12)
	version := binary.BigEndian.Uint32(header[12:16])

	if version < 5 {
		return nil, fmt.Errorf("CHD version %d not supported (only v5+ supported)", version)
	}

	if headerLen < headerSize {
		return nil, fmt.Errorf("CHD header too small: %d bytes", headerLen)
	}

	// Compressors (4 x uint32 at offsets 16-31)
	var compressors [4]uint32
	for i := 0; i < 4; i++ {
		compressors[i] = binary.BigEndian.Uint32(header[16+i*4:])
	}

	// Logical bytes (uint64 at offset 32)
	logicalBytes := binary.BigEndian.Uint64(header[32:40])

	// Map offset (uint64 at offset 40)
	mapOffset := binary.BigEndian.Uint64(header[40:48])

	// Metadata offset (uint64 at offset 48)
	metaOffset := binary.BigEndian.Uint64(header[48:56])

	// Hunk bytes (uint32 at offset 56)
	hunkBytes := binary.BigEndian.Uint32(header[56:60])

	// Unit bytes (uint32 at offset 60)
	unitBytes := binary.BigEndian.Uint32(header[60:64])

	// Calculate total hunks
	var totalHunks uint32
	if hunkBytes > 0 {
		totalHunks = uint32((logicalBytes + uint64(hunkBytes) - 1) / uint64(hunkBytes))
	}

	// Raw SHA1 at offset 64 (20 bytes) - SHA1 of the uncompressed data
	rawSHA1 := hex.EncodeToString(header[rawSHA1Offset : rawSHA1Offset+sha1Size])

	// SHA1 at offset 84 (20 bytes) - SHA1 of the compressed data
	sha1 := hex.EncodeToString(header[sha1Offset : sha1Offset+sha1Size])

	// Parent SHA1 at offset 104 (20 bytes) - all zeros if no parent
	parentSHA1Bytes := header[parentSHA1Offset : parentSHA1Offset+sha1Size]
	parentSHA1 := ""
	for _, b := range parentSHA1Bytes {
		if b != 0 {
			parentSHA1 = hex.EncodeToString(parentSHA1Bytes)
			break
		}
	}

	return &Header{
		Version:      version,
		Compressors:  compressors,
		LogicalBytes: logicalBytes,
		MapOffset:    mapOffset,
		MetaOffset:   metaOffset,
		HunkBytes:    hunkBytes,
		UnitBytes:    unitBytes,
		TotalHunks:   totalHunks,
		RawSHA1:      rawSHA1,
		SHA1:         sha1,
		ParentSHA1:   parentSHA1,
	}, nil
}

// IsCompressed returns true if the CHD uses compression (has at least one codec).
func (h *Header) IsCompressed() bool {
	return h.Compressors[0] != CodecNone
}

// IsCDROM returns true if this appears to be a CD-ROM image based on unit size.
func (h *Header) IsCDROM() bool {
	// CD-ROM CHDs typically use 2448 bytes per unit (2352 sector + 96 subcode)
	// or 2352 bytes per unit
	return h.UnitBytes == 2448 || h.UnitBytes == 2352
}
