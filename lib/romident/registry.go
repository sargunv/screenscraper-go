package romident

import (
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/romident/game"
	"github.com/sargunv/rom-tools/lib/romident/game/gb"
	"github.com/sargunv/rom-tools/lib/romident/game/gba"
	"github.com/sargunv/rom-tools/lib/romident/game/md"
	"github.com/sargunv/rom-tools/lib/romident/game/n64"
	"github.com/sargunv/rom-tools/lib/romident/game/nds"
	"github.com/sargunv/rom-tools/lib/romident/game/nes"
	"github.com/sargunv/rom-tools/lib/romident/game/snes"
	"github.com/sargunv/rom-tools/lib/romident/game/xbox"
)

// FormatEntry associates a format with its extensions and identification function.
type FormatEntry struct {
	Format     Format
	Extensions []string
	Identify   game.IdentifyFunc
}

// registry contains all registered format entries.
// Entries are ordered by specificity - more specific extensions first.
var registry = []FormatEntry{
	// GBA
	{FormatGBA, []string{".gba"}, gba.Identify},
	// Nintendo DS
	{FormatNDS, []string{".nds", ".dsi", ".ids"}, nds.Identify},
	// NES
	{FormatNES, []string{".nes"}, nes.Identify},
	// SNES
	{FormatSNES, []string{".sfc", ".smc"}, snes.Identify},
	// Game Boy
	{FormatGB, []string{".gb", ".gbc"}, gb.Identify},
	// N64 variants
	{FormatZ64, []string{".z64"}, n64.IdentifyZ64},
	{FormatV64, []string{".v64"}, n64.IdentifyV64},
	{FormatN64, []string{".n64"}, n64.IdentifyN64},
	// Mega Drive / Genesis
	{FormatMD, []string{".md", ".gen"}, md.Identify},
	{FormatSMD, []string{".smd"}, md.IdentifySMD},
	// Xbox
	{FormatXISO, []string{".xiso"}, xbox.IdentifyXISO},
	{FormatXBE, []string{".xbe"}, xbox.IdentifyXBE},

	// Container/disc formats without game identifiers
	{FormatCHD, []string{".chd"}, nil},
	{FormatZIP, []string{".zip"}, nil},
	{FormatISO9660, []string{".iso"}, nil},
	// Note: .iso can also be XISO, handled by trying multiple candidates
}

// ambiguousExtensions maps extensions that can match multiple formats.
var ambiguousExtensions = map[string][]Format{
	".iso": {FormatXISO, FormatISO9660},
}

// FormatsByExtension returns all format entries that match the given filename extension.
// Returns entries in order of preference (more specific formats first).
func FormatsByExtension(filename string) []FormatEntry {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil
	}

	// Check for ambiguous extensions first
	if formats, ok := ambiguousExtensions[ext]; ok {
		var entries []FormatEntry
		for _, f := range formats {
			for _, entry := range registry {
				if entry.Format == f {
					entries = append(entries, entry)
					break
				}
			}
		}
		return entries
	}

	// Find matching entries
	var entries []FormatEntry
	for _, entry := range registry {
		for _, entryExt := range entry.Extensions {
			if entryExt == ext {
				entries = append(entries, entry)
				break
			}
		}
	}
	return entries
}
