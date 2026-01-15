package chd

import (
	"io"

	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/iso9660"
)

// Identify attempts to identify the game contained in a CHD file.
// It decompresses sectors as needed to extract ISO9660 filesystem data,
// then delegates to platform-specific identification routines.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	// Create CHD reader
	reader, err := NewReader(r)
	if err != nil {
		return nil, err
	}

	// Wrap CHD reader to provide ISO9660-compatible sector access
	var isoReader io.ReaderAt
	var isoSize int64

	if reader.Header().IsCDROM() {
		// CD-ROM CHD: sectors are 2448 bytes (2352 data + 96 subcode)
		// User data is at offset 16 within the 2352-byte sector data
		isoReader = &chdSectorReader{
			reader:     reader,
			sectorSize: int64(reader.Header().UnitBytes),
			dataOffset: 16, // Mode 1 data offset within sector
		}
		// Calculate logical size (number of sectors * 2048)
		numSectors := int64(reader.Header().LogicalBytes) / int64(reader.Header().UnitBytes)
		isoSize = numSectors * 2048
	} else {
		// DVD/other CHD: sectors are typically 2048 bytes (no translation needed)
		isoReader = reader
		isoSize = int64(reader.Header().LogicalBytes)
	}

	// Delegate to ISO9660 identification
	return iso9660.Identify(isoReader, isoSize)
}

// chdSectorReader translates logical 2048-byte sector reads to CHD raw sector reads.
// For CD-ROM CHDs, it extracts user data from raw 2352-byte sectors.
type chdSectorReader struct {
	reader     *Reader
	sectorSize int64 // Physical sector size (typically 2448 for CD-ROM)
	dataOffset int64 // Offset to user data within physical sector (16 for Mode 1)
}

// ReadAt implements io.ReaderAt, reading logical data from CHD sectors.
func (c *chdSectorReader) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	for n < len(p) {
		logicalOffset := off + int64(n)

		// Which logical sector (2048-byte)?
		logicalSector := logicalOffset / 2048
		offsetInSector := logicalOffset % 2048

		// Read the physical sector from CHD
		sectorData, err := c.reader.ReadSector(uint64(logicalSector))
		if err != nil {
			if n > 0 {
				return n, nil
			}
			return 0, err
		}

		// Extract user data starting at dataOffset
		// For CD-ROM Mode 1: bytes 16-2063 of the 2352-byte sector
		dataStart := int(c.dataOffset + offsetInSector)
		dataEnd := int(c.dataOffset) + 2048

		if dataStart >= len(sectorData) {
			if n > 0 {
				return n, nil
			}
			return 0, io.EOF
		}

		if dataEnd > len(sectorData) {
			dataEnd = len(sectorData)
		}

		// Copy available data
		bytesToCopy := dataEnd - dataStart
		if bytesToCopy > len(p)-n {
			bytesToCopy = len(p) - n
		}

		copy(p[n:n+bytesToCopy], sectorData[dataStart:dataStart+bytesToCopy])
		n += bytesToCopy
	}

	return n, nil
}
