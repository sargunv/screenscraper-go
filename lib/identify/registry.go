package identify

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gb"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gba"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/gcm"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/n3ds"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/n64"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/nds"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/nes"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/rvz"
	"github.com/sargunv/rom-tools/lib/roms/nintendo/sfc"
	"github.com/sargunv/rom-tools/lib/roms/playstation/pkg"
	"github.com/sargunv/rom-tools/lib/roms/sega/md"
	"github.com/sargunv/rom-tools/lib/roms/sega/sms"
	"github.com/sargunv/rom-tools/lib/roms/xbox/xbe"
	"github.com/sargunv/rom-tools/lib/roms/xbox/xiso"
)

// identifyFunc attempts to identify content from a reader.
// Returns game info, optional embedded hashes (for formats like CHD), and error.
type identifyFunc func(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error)

// wrapParser converts a typed parser function to the generic signature.
// This is needed because Go function types are invariant - a function returning
// *GBAInfo is not assignable to a function returning GameInfo even though
// *GBAInfo implements GameInfo.
func wrapParser[T core.GameInfo](fn func(io.ReaderAt, int64) (T, error)) identifyFunc {
	return func(r io.ReaderAt, size int64) (core.GameInfo, core.Hashes, error) {
		info, err := fn(r, size)
		return info, nil, err
	}
}

// registry maps file extensions to ordered list of parsers to try.
// Parsers are tried in order until one succeeds.
var registry = map[string][]identifyFunc{
	".gba":  {wrapParser(gba.Parse)},
	".gb":   {wrapParser(gb.Parse)},
	".gbc":  {wrapParser(gb.Parse)},
	".nds":  {wrapParser(nds.Parse)},
	".dsi":  {wrapParser(nds.Parse)},
	".ids":  {wrapParser(nds.Parse)},
	".3ds":  {wrapParser(n3ds.Parse)},
	".cci":  {wrapParser(n3ds.Parse)},
	".nes":  {wrapParser(nes.Parse)},
	".sfc":  {wrapParser(sfc.Parse)},
	".smc":  {wrapParser(sfc.Parse)},
	".z64":  {wrapParser(n64.Parse)},
	".v64":  {wrapParser(n64.Parse)},
	".n64":  {wrapParser(n64.Parse)},
	".md":   {wrapParser(md.Parse)},
	".gen":  {wrapParser(md.Parse)},
	".32x":  {wrapParser(md.Parse)},
	".smd":  {wrapParser(md.Parse)},
	".sms":  {wrapParser(sms.Parse)},
	".gg":   {wrapParser(sms.Parse)},
	".xbe":  {wrapParser(xbe.Parse)},
	".pkg":  {wrapParser(pkg.Parse)},
	".chd":  {identifyCHD},
	".rvz":  {wrapParser(rvz.Parse)},
	".wia":  {wrapParser(rvz.Parse)},
	".gcm":  {wrapParser(gcm.Parse)},
	".xiso": {wrapParser(xiso.Parse)},
	".iso":  {wrapParser(xiso.Parse), wrapParser(gcm.Parse), identifyISO9660},
	".bin":  {identifyISO9660, wrapParser(md.Parse)},
}

// identifyByExtension returns the list of parsers to try for a given filename.
func identifyByExtension(filename string) []identifyFunc {
	ext := strings.ToLower(filepath.Ext(filename))
	return registry[ext]
}
