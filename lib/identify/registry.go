package identify

import (
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sargunv/rom-tools/lib/roms/gamecube"
	"github.com/sargunv/rom-tools/lib/roms/gb"
	"github.com/sargunv/rom-tools/lib/roms/gba"
	"github.com/sargunv/rom-tools/lib/roms/megadrive"
	"github.com/sargunv/rom-tools/lib/roms/n64"
	"github.com/sargunv/rom-tools/lib/roms/nds"
	"github.com/sargunv/rom-tools/lib/roms/nes"
	"github.com/sargunv/rom-tools/lib/roms/sms"
	"github.com/sargunv/rom-tools/lib/roms/snes"
	"github.com/sargunv/rom-tools/lib/roms/xbox"
)

// formatEntry associates a format with its extensions and identification function.
// TODO: Format should be returned by parsers via GameInfo (like Platform) rather than
// being defined here. This would allow removing the Format field from the registry.
type formatEntry struct {
	Format     Format
	Extensions []string
	Identify   func(r io.ReaderAt, size int64) (GameInfo, error)
}

// registry contains all registered format entries.
// Entries are ordered by specificity - more specific extensions first.
// For ambiguous extensions like .iso, multiple formats are registered and
// the detection logic tries each candidate in order.
var registry = []formatEntry{
	{FormatGBA, []string{".gba"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return gba.ParseGBA(r, size) }},
	{FormatNDS, []string{".nds", ".dsi", ".ids"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return nds.ParseNDS(r, size) }},
	{FormatNES, []string{".nes"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return nes.ParseNES(r, size) }},
	{FormatSNES, []string{".sfc", ".smc"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return snes.ParseSNES(r, size) }},
	{FormatGB, []string{".gb", ".gbc"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return gb.ParseGB(r, size) }},
	{FormatZ64, []string{".z64"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return n64.ParseN64(r, size) }},
	{FormatV64, []string{".v64"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return n64.ParseN64(r, size) }},
	{FormatN64, []string{".n64"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return n64.ParseN64(r, size) }},
	{FormatMD, []string{".32x", ".md", ".gen"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return megadrive.Parse(r, size) }},
	{FormatSMD, []string{".smd"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return megadrive.Parse(r, size) }},
	{FormatSMS, []string{".sms", ".gg"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return sms.ParseSMS(r, size) }},
	{FormatXISO, []string{".xiso", ".iso"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return xbox.ParseXISO(r, size) }},
	{FormatXBE, []string{".xbe"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return xbox.ParseXBE(r, size) }},
	{FormatGCM, []string{".gcm", ".iso"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return gamecube.ParseGCM(r, size) }},
	{FormatRVZ, []string{".rvz", ".wia"}, func(r io.ReaderAt, size int64) (GameInfo, error) { return gamecube.ParseRVZ(r, size) }},
	{FormatCHD, []string{".chd"}, identifyCHD},
	{FormatZIP, []string{".zip"}, nil},
	{FormatISO9660, []string{".iso", ".bin"}, identifyISO9660},
}

// formatsByExtension returns all format entries that match the given filename extension.
// Returns entries in order of preference (more specific formats first).
func formatsByExtension(filename string) []formatEntry {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil
	}

	// Find all matching entries
	var entries []formatEntry
	for _, entry := range registry {
		if slices.Contains(entry.Extensions, ext) {
			entries = append(entries, entry)
		}
	}
	return entries
}
