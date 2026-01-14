package sfo

import (
	"encoding/binary"
	"fmt"
)

// SFO (System File Object) binary format parser.
//
// SFO is a metadata format used by PlayStation platforms (PSP, PS3, PS Vita)
// to store game information like title, disc ID, and system requirements.
//
// Format structure:
//   - Header (20 bytes): magic, version, table offsets, entry count
//   - Index table: 16-byte entries describing each key-value pair
//   - Key table: null-terminated UTF-8 key names
//   - Data table: values (strings or integers)
//
// References:
//   - https://www.psdevwiki.com/psp/Param.sfo
//   - https://www.psdevwiki.com/ps3/PARAM.SFO

const (
	sfoMagic = "\x00PSF"

	// Data format types
	formatUTF8Special = 0x0004 // UTF-8 not null-terminated
	formatUTF8        = 0x0204 // UTF-8 null-terminated
	formatInt32       = 0x0404 // 32-bit unsigned integer
)

// Parse reads an SFO file and returns key-value pairs.
// String values are returned as string, integers as uint32.
func Parse(data []byte) (map[string]any, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("SFO too small: %d bytes", len(data))
	}

	// Validate magic
	if string(data[0:4]) != sfoMagic {
		return nil, fmt.Errorf("invalid SFO magic: %x", data[0:4])
	}

	// Read header
	keyTableOffset := binary.LittleEndian.Uint32(data[8:12])
	dataTableOffset := binary.LittleEndian.Uint32(data[12:16])
	numEntries := binary.LittleEndian.Uint32(data[16:20])

	// Validate offsets
	if keyTableOffset > uint32(len(data)) || dataTableOffset > uint32(len(data)) {
		return nil, fmt.Errorf("SFO table offsets out of bounds")
	}

	result := make(map[string]any)

	// Parse index entries (16 bytes each, starting at offset 20)
	indexOffset := uint32(20)
	for i := uint32(0); i < numEntries; i++ {
		entryOffset := indexOffset + i*16
		if entryOffset+16 > uint32(len(data)) {
			return nil, fmt.Errorf("SFO index entry %d out of bounds", i)
		}

		keyOffset := binary.LittleEndian.Uint16(data[entryOffset:])
		dataFormat := binary.LittleEndian.Uint16(data[entryOffset+2:])
		dataLen := binary.LittleEndian.Uint32(data[entryOffset+4:])
		dataOffset := binary.LittleEndian.Uint32(data[entryOffset+12:])

		// Read key name (null-terminated string)
		keyStart := keyTableOffset + uint32(keyOffset)
		if keyStart >= uint32(len(data)) {
			return nil, fmt.Errorf("SFO key %d offset out of bounds", i)
		}
		keyEnd := keyStart
		for keyEnd < uint32(len(data)) && data[keyEnd] != 0 {
			keyEnd++
		}
		key := string(data[keyStart:keyEnd])

		// Read data value
		dataStart := dataTableOffset + dataOffset
		if dataStart+dataLen > uint32(len(data)) {
			return nil, fmt.Errorf("SFO data for key %q out of bounds", key)
		}

		switch dataFormat {
		case formatInt32:
			if dataLen >= 4 {
				result[key] = binary.LittleEndian.Uint32(data[dataStart:])
			}
		case formatUTF8, formatUTF8Special:
			// Trim null terminator if present
			strData := data[dataStart : dataStart+dataLen]
			for len(strData) > 0 && strData[len(strData)-1] == 0 {
				strData = strData[:len(strData)-1]
			}
			result[key] = string(strData)
		default:
			// Unknown format, store as raw bytes
			result[key] = data[dataStart : dataStart+dataLen]
		}
	}

	return result, nil
}

// GetString returns a string value from parsed SFO data.
func GetString(sfo map[string]any, key string) string {
	if v, ok := sfo[key].(string); ok {
		return v
	}
	return ""
}

// GetUint32 returns a uint32 value from parsed SFO data.
func GetUint32(sfo map[string]any, key string) uint32 {
	if v, ok := sfo[key].(uint32); ok {
		return v
	}
	return 0
}
