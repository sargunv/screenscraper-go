package ps2

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/core"
)

// PS2 SYSTEM.CNF parsing for disc identification.
//
// PS2 discs contain a SYSTEM.CNF file at the root with format:
//
//	BOOT2 = cdrom0:\SLUS_123.45;1
//	VER = 1.00
//	VMODE = NTSC
//
// The disc ID (e.g., "SLUS_123.45") is the executable filename.
// The prefix encodes the region:
//   - SLUS/SCUS = US (NTSC-U)
//   - SLES/SCES = EU (PAL)
//   - SLPS/SCPS/SLPM = JP (NTSC-J)
//   - SLKA/SCKA = KR

// PS2Info contains metadata extracted from a PS2 disc.
type PS2Info struct {
	DiscID    string // e.g., "SLUS_123.45"
	Version   string // from VER line
	VideoMode string // NTSC or PAL
}

// IdentifyFromSystemCNF parses PS2 SYSTEM.CNF content.
// Returns nil if not a valid PS2 SYSTEM.CNF (e.g., PS1 format uses BOOT instead of BOOT2).
func IdentifyFromSystemCNF(data []byte) *core.GameIdent {
	info := parseSystemCNF(data)
	if info == nil {
		return nil
	}

	return &core.GameIdent{
		Platform: core.PlatformPS2,
		TitleID:  info.DiscID,
		Regions:  []core.Region{decodeRegion(info.DiscID)},
		Extra:    info,
	}
}

// parseSystemCNF extracts PS2 info from SYSTEM.CNF content.
// Returns nil if BOOT2 line is not found (indicates non-PS2, e.g., PS1).
func parseSystemCNF(data []byte) *PS2Info {
	info := &PS2Info{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Split on '=' and trim spaces
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "BOOT2":
			// Extract disc ID from path: cdrom0:\SLUS_123.45;1
			info.DiscID = extractDiscID(value)
		case "VER":
			info.Version = value
		case "VMODE":
			info.VideoMode = value
		}
	}

	// BOOT2 is required for PS2 identification
	if info.DiscID == "" {
		return nil
	}

	return info
}

// extractDiscID extracts the disc ID from a BOOT2 path.
// Input: "cdrom0:\SLUS_123.45;1" or similar
// Output: "SLUS_123.45"
func extractDiscID(bootPath string) string {
	// Find last backslash
	lastBackslash := strings.LastIndex(bootPath, "\\")
	if lastBackslash == -1 {
		// Try forward slash as fallback
		lastBackslash = strings.LastIndex(bootPath, "/")
	}

	var filename string
	if lastBackslash != -1 {
		filename = bootPath[lastBackslash+1:]
	} else {
		filename = bootPath
	}

	// Remove version suffix (;1)
	if semicolon := strings.Index(filename, ";"); semicolon != -1 {
		filename = filename[:semicolon]
	}

	return filename
}

// decodeRegion maps a PS2 disc ID prefix to a region.
func decodeRegion(discID string) core.Region {
	if len(discID) < 4 {
		return core.RegionUnknown
	}

	prefix := strings.ToUpper(discID[:4])

	switch prefix {
	case "SLUS", "SCUS":
		return core.RegionUS
	case "SLES", "SCES":
		return core.RegionEU
	case "SLPS", "SCPS", "SLPM", "SCPM":
		return core.RegionJP
	case "SLKA", "SCKA":
		return core.RegionKR
	case "SLAJ": // Asia/Japan variant
		return core.RegionJP
	default:
		return core.RegionUnknown
	}
}
