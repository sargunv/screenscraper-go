package chd

import (
	"fmt"
	"io"
	"sync"
)

// Reader provides sector-level access to CHD files.
type Reader struct {
	file      io.ReaderAt
	header    *Header
	hunkMap   *Map
	hunkCache map[uint32][]byte
	cacheMu   sync.RWMutex
}

// NewReader creates a new CHD reader from a file.
func NewReader(r io.ReaderAt) (*Reader, error) {
	// Parse header
	header, err := ParseCHDHeader(r)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	// Check for unsupported features
	if header.ParentSHA1 != "" {
		return nil, fmt.Errorf("parent CHD references not supported")
	}

	// Decode hunk map
	hunkMap, err := DecodeMap(r, header)
	if err != nil {
		return nil, fmt.Errorf("decode hunk map: %w", err)
	}

	return &Reader{
		file:      r,
		header:    header,
		hunkMap:   hunkMap,
		hunkCache: make(map[uint32][]byte),
	}, nil
}

// Header returns the CHD header information.
func (r *Reader) Header() *Header {
	return r.header
}

// Size returns the logical (uncompressed) size of the CHD.
func (r *Reader) Size() int64 {
	return int64(r.header.LogicalBytes)
}

// ReadAt implements io.ReaderAt, reading from the logical (uncompressed) data.
func (r *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, fmt.Errorf("negative offset")
	}
	if off >= int64(r.header.LogicalBytes) {
		return 0, io.EOF
	}

	hunkBytes := int64(r.header.HunkBytes)
	remaining := len(p)
	pos := off

	for remaining > 0 && pos < int64(r.header.LogicalBytes) {
		// Calculate which hunk contains this position
		hunkNum := uint32(pos / hunkBytes)
		hunkOffset := int(pos % hunkBytes)

		// Get the decompressed hunk
		hunkData, err := r.readHunk(hunkNum)
		if err != nil {
			if n > 0 {
				return n, nil // Return partial read
			}
			return 0, fmt.Errorf("read hunk %d: %w", hunkNum, err)
		}

		// Copy data from this hunk
		available := len(hunkData) - hunkOffset
		if available <= 0 {
			break
		}
		toCopy := remaining
		if toCopy > available {
			toCopy = available
		}

		copy(p[n:n+toCopy], hunkData[hunkOffset:hunkOffset+toCopy])
		n += toCopy
		remaining -= toCopy
		pos += int64(toCopy)
	}

	if n == 0 && remaining > 0 {
		return 0, io.EOF
	}
	return n, nil
}

// readHunk reads and decompresses a single hunk.
func (r *Reader) readHunk(hunkNum uint32) ([]byte, error) {
	// Check cache first
	r.cacheMu.RLock()
	if cached, ok := r.hunkCache[hunkNum]; ok {
		r.cacheMu.RUnlock()
		return cached, nil
	}
	r.cacheMu.RUnlock()

	if int(hunkNum) >= len(r.hunkMap.Entries) {
		return nil, fmt.Errorf("hunk %d out of range (total: %d)", hunkNum, len(r.hunkMap.Entries))
	}

	entry := r.hunkMap.Entries[hunkNum]
	hunkBytes := r.header.HunkBytes

	var data []byte
	var err error

	switch entry.Compression {
	case compressionNone:
		// Uncompressed - read directly from file
		data = make([]byte, hunkBytes)
		_, err = r.file.ReadAt(data, int64(entry.Offset))
		if err != nil {
			return nil, fmt.Errorf("read uncompressed hunk: %w", err)
		}

	case compressionType0, compressionType1, compressionType2, compressionType3:
		// Compressed - use the corresponding compressor
		codecID := r.header.Compressors[entry.Compression]

		compressed := make([]byte, entry.Length)
		_, err = r.file.ReadAt(compressed, int64(entry.Offset))
		if err != nil {
			return nil, fmt.Errorf("read compressed data: %w", err)
		}

		data, err = decompressHunk(compressed, codecID, hunkBytes)
		if err != nil {
			return nil, fmt.Errorf("decompress hunk (codec 0x%08x): %w", codecID, err)
		}

	case compressionSelf:
		// Self-reference - copy from another hunk
		refHunk := uint32(entry.Offset)
		if refHunk >= hunkNum {
			return nil, fmt.Errorf("self-reference to hunk %d from hunk %d (forward reference)", refHunk, hunkNum)
		}
		data, err = r.readHunk(refHunk)
		if err != nil {
			return nil, fmt.Errorf("read self-referenced hunk %d: %w", refHunk, err)
		}
		// Make a copy so cache entries don't share backing arrays
		data = append([]byte(nil), data...)

	case compressionParent:
		return nil, fmt.Errorf("parent CHD references not supported")

	default:
		return nil, fmt.Errorf("unknown compression type: %d", entry.Compression)
	}

	// Cache the decompressed hunk (limit cache size to avoid memory issues)
	r.cacheMu.Lock()
	if len(r.hunkCache) < 32 { // Cache up to 32 hunks
		r.hunkCache[hunkNum] = data
	}
	r.cacheMu.Unlock()

	return data, nil
}

// ReadSector reads a single sector (unit) from the CHD.
// For CD-ROM, this returns 2352 bytes (raw frame) or 2448 bytes (frame + subcode).
// For DVD/other, this returns the unit size specified in the header.
func (r *Reader) ReadSector(sectorNum uint64) ([]byte, error) {
	unitBytes := uint64(r.header.UnitBytes)
	offset := int64(sectorNum * unitBytes)

	data := make([]byte, unitBytes)
	n, err := r.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return data[:n], nil
}

// ReadSectors reads multiple consecutive sectors.
func (r *Reader) ReadSectors(start, count uint64) ([]byte, error) {
	unitBytes := uint64(r.header.UnitBytes)
	offset := int64(start * unitBytes)
	size := count * unitBytes

	data := make([]byte, size)
	n, err := r.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return data[:n], nil
}
