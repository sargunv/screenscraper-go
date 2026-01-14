package iso9660

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/cnf"
	"github.com/sargunv/rom-tools/lib/romident/core"
	"github.com/sargunv/rom-tools/lib/romident/psp"
)

// ISO 9660 disc identification with platform dispatch.
//
// This package handles ISO 9660 filesystem parsing and dispatches to
// platform-specific handlers based on the disc contents:
//   - SYSTEM.CNF with BOOT2 → PS2
//   - SYSTEM.CNF with BOOT → PS1
//   - PSP_GAME/PARAM.SFO → PSP
//
// ISO 9660 layout (relevant parts):
//   - Sector 16 (offset 0x8000): Primary Volume Descriptor
//   - PVD offset 156: Root directory record (34 bytes)
//   - Root directory record contains extent location and size

const (
	sectorSize        = 2048
	pvdOffset         = 16 * sectorSize // Sector 16
	pvdMagicOffset    = 1
	pvdRootDirOffset  = 156
	dirEntryExtentLoc = 2  // Offset within directory entry
	dirEntryDataLen   = 10 // Offset within directory entry
	dirEntryFlags     = 25 // Offset within directory entry (bit 1 = directory)
	dirEntryNameLen   = 32 // Offset within directory entry
	dirEntryName      = 33 // Offset within directory entry

	flagDirectory = 0x02 // Directory flag in file flags byte
)

// Identify parses an ISO 9660 image and attempts to identify the platform.
// Returns nil (not an error) if the ISO is valid but the platform is unknown.
func Identify(r io.ReaderAt, size int64) (*core.GameIdent, error) {
	img, err := openImage(r, size)
	if err != nil {
		return nil, err
	}

	// Try to read SYSTEM.CNF (PS1/PS2 discs)
	if data, err := img.readFile("SYSTEM.CNF"); err == nil {
		if ident := cnf.IdentifyFromSystemCNF(data); ident != nil {
			return ident, nil
		}
	}

	// Try to read PSP_GAME/PARAM.SFO (PSP discs)
	if data, err := img.readFile("PSP_GAME/PARAM.SFO"); err == nil {
		if ident := psp.IdentifyFromSFO(data); ident != nil {
			return ident, nil
		}
	}

	// No platform identified - valid ISO but unknown content
	return nil, nil
}

// image represents a minimal ISO 9660 image reader.
type image struct {
	r             io.ReaderAt
	rootExtentLoc uint32
	rootExtentLen uint32
}

// openImage opens an ISO 9660 image and validates the primary volume descriptor.
func openImage(r io.ReaderAt, size int64) (*image, error) {
	if size < pvdOffset+sectorSize {
		return nil, fmt.Errorf("file too small for ISO 9660")
	}

	// Read Primary Volume Descriptor
	pvd := make([]byte, sectorSize)
	if _, err := r.ReadAt(pvd, pvdOffset); err != nil {
		return nil, fmt.Errorf("failed to read PVD: %w", err)
	}

	// Validate magic "CD001"
	if string(pvd[pvdMagicOffset:pvdMagicOffset+5]) != "CD001" {
		return nil, fmt.Errorf("not a valid ISO 9660: missing CD001 magic")
	}

	// Extract root directory record info
	rootRecord := pvd[pvdRootDirOffset:]
	rootExtentLoc := binary.LittleEndian.Uint32(rootRecord[dirEntryExtentLoc:])
	rootExtentLen := binary.LittleEndian.Uint32(rootRecord[dirEntryDataLen:])

	return &image{
		r:             r,
		rootExtentLoc: rootExtentLoc,
		rootExtentLen: rootExtentLen,
	}, nil
}

// readFile reads a file by path (case-insensitive).
// Supports subdirectory paths like "PSP_GAME/PARAM.SFO".
// Handles ISO 9660 version suffixes (e.g., ";1").
func (img *image) readFile(path string) ([]byte, error) {
	// Split path into components
	parts := strings.Split(path, "/")

	// Start from root directory
	dirExtentLoc := img.rootExtentLoc
	dirExtentLen := img.rootExtentLen

	// Traverse directories
	for i, part := range parts {
		isLast := i == len(parts)-1

		extentLoc, extentLen, isDir, err := img.findEntry(dirExtentLoc, dirExtentLen, part)
		if err != nil {
			return nil, fmt.Errorf("path component %q not found: %w", part, err)
		}

		if isLast {
			// Final component - read the file
			if isDir {
				return nil, fmt.Errorf("%q is a directory, not a file", part)
			}
			data := make([]byte, extentLen)
			if _, err := img.r.ReadAt(data, int64(extentLoc)*sectorSize); err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			return data, nil
		}

		// Intermediate component - must be a directory
		if !isDir {
			return nil, fmt.Errorf("%q is not a directory", part)
		}
		dirExtentLoc = extentLoc
		dirExtentLen = extentLen
	}

	return nil, fmt.Errorf("empty path")
}

// findEntry searches a directory for an entry by name.
// Returns the entry's extent location, size, whether it's a directory, and any error.
func (img *image) findEntry(dirExtentLoc, dirExtentLen uint32, name string) (uint32, uint32, bool, error) {
	// Read directory
	dirData := make([]byte, dirExtentLen)
	if _, err := img.r.ReadAt(dirData, int64(dirExtentLoc)*sectorSize); err != nil {
		return 0, 0, false, fmt.Errorf("failed to read directory: %w", err)
	}

	name = strings.ToUpper(name)
	offset := 0
	for offset < len(dirData) {
		entryLen := int(dirData[offset])
		if entryLen == 0 {
			// End of directory entries in this sector, try next sector
			nextSector := ((offset / sectorSize) + 1) * sectorSize
			if nextSector >= len(dirData) {
				break
			}
			offset = nextSector
			continue
		}

		if offset+dirEntryName >= len(dirData) {
			break
		}

		nameLen := int(dirData[offset+dirEntryNameLen])
		if offset+dirEntryName+nameLen > len(dirData) {
			break
		}

		entryName := strings.ToUpper(string(dirData[offset+dirEntryName : offset+dirEntryName+nameLen]))

		// Strip version suffix (";1")
		if idx := strings.Index(entryName, ";"); idx != -1 {
			entryName = entryName[:idx]
		}

		if entryName == name {
			extentLoc := binary.LittleEndian.Uint32(dirData[offset+dirEntryExtentLoc:])
			extentLen := binary.LittleEndian.Uint32(dirData[offset+dirEntryDataLen:])
			flags := dirData[offset+dirEntryFlags]
			isDir := (flags & flagDirectory) != 0
			return extentLoc, extentLen, isDir, nil
		}

		offset += entryLen
	}

	return 0, 0, false, fmt.Errorf("entry not found: %s", name)
}
