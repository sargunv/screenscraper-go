package gc

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/romident/disc/rvz"
	"github.com/sargunv/rom-tools/lib/romident/game"
)

// GameCube/Wii disc header format parsing.
//
// Both GameCube and Wii use the same disc header format.
// Per WiiBrew wiki: "The first 0x400 bytes are like the GameCube disc header format."
// https://wiibrew.org/wiki/Wii_disc
//
// Disc header layout (relevant fields):
//
//	Offset  Size  Description
//	0x000   1     Disc ID (G=GameCube, R/S=Wii, D=GC Demo)
//	0x001   2     Game code
//	0x003   1     Region code (E=USA, P=PAL, J=Japan, etc.)
//	0x004   2     Maker code
//	0x006   1     Disc number
//	0x007   1     Disc version
//	0x018   4     Wii magic word (0x5D1C9EA3 for Wii, 0x00000000 for GameCube)
//	0x01C   4     GameCube magic word (0xC2339F3D for GameCube, 0x00000000 for Wii)
//	0x020   64    Game title (ASCII, null-terminated)

const (
	discHeaderSize = 0x60 // We only need first 96 bytes for identification

	discIDOffset      = 0x000
	gameCodeOffset    = 0x001
	gameCodeLen       = 2
	regionCodeOffset  = 0x003
	makerCodeOffset   = 0x004
	makerCodeLen      = 2
	discNumberOffset  = 0x006
	discVersionOffset = 0x007
	wiiMagicOffset    = 0x018
	gcMagicOffset     = 0x01C
	titleOffset       = 0x020
	titleLen          = 64

	wiiMagicWord = 0x5D1C9EA3
	gcMagicWord  = 0xC2339F3D
)

// GCInfo contains metadata extracted from a GameCube/Wii disc header.
type GCInfo struct {
	DiscID     byte   // Console identifier (G=GameCube, R/S=Wii, D=GC Demo)
	GameCode   string // 2-character unique game identifier
	RegionCode byte   // E=USA, P=PAL, J=Japan, etc.
	MakerCode  string // 2-character publisher identifier
	DiscNumber int    // Disc number for multi-disc games
	Version    int    // Disc version/revision
	Title      string // Game title
	IsWii      bool   // True if Wii disc, false if GameCube
}

// ParseDiscHeader parses a GameCube/Wii disc header from raw bytes.
// Requires at least 0x60 bytes of header data.
func ParseDiscHeader(header []byte) (*GCInfo, error) {
	if len(header) < discHeaderSize {
		return nil, fmt.Errorf("header too small: need %d bytes, got %d", discHeaderSize, len(header))
	}

	// Check magic words to determine platform and validate
	wiiMagic := binary.BigEndian.Uint32(header[wiiMagicOffset:])
	gcMagic := binary.BigEndian.Uint32(header[gcMagicOffset:])

	isWii := wiiMagic == wiiMagicWord
	isGC := gcMagic == gcMagicWord

	if !isWii && !isGC {
		return nil, fmt.Errorf("not a valid GameCube/Wii disc: no magic word found (Wii: 0x%08X, GC: 0x%08X)",
			wiiMagic, gcMagic)
	}

	// Extract fields
	discID := header[discIDOffset]
	gameCode := util.ExtractASCII(header[gameCodeOffset : gameCodeOffset+gameCodeLen])
	regionCode := header[regionCodeOffset]
	makerCode := util.ExtractASCII(header[makerCodeOffset : makerCodeOffset+makerCodeLen])
	discNumber := int(header[discNumberOffset])
	version := int(header[discVersionOffset])
	title := util.ExtractASCII(header[titleOffset : titleOffset+titleLen])

	return &GCInfo{
		DiscID:     discID,
		GameCode:   gameCode,
		RegionCode: regionCode,
		MakerCode:  makerCode,
		DiscNumber: discNumber,
		Version:    version,
		Title:      title,
		IsWii:      isWii,
	}, nil
}

// IdentifyGCM verifies the format and extracts game identification from a GCM/ISO file.
func IdentifyGCM(r io.ReaderAt, size int64) (*game.GameIdent, error) {
	if size < discHeaderSize {
		return nil, fmt.Errorf("file too small for disc header: %d bytes", size)
	}

	header := make([]byte, discHeaderSize)
	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("failed to read disc header: %w", err)
	}

	info, err := ParseDiscHeader(header)
	if err != nil {
		return nil, err
	}

	return gcInfoToGameIdent(info, nil), nil
}

// IdentifyRVZ verifies the format and extracts game identification from an RVZ/WIA file.
func IdentifyRVZ(r io.ReaderAt, size int64) (*game.GameIdent, error) {
	rvzInfo, err := rvz.ParseRVZHeader(r, size)
	if err != nil {
		return nil, err
	}

	gcInfo, err := ParseDiscHeader(rvzInfo.DiscHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse disc header from RVZ: %w", err)
	}

	return gcInfoToGameIdent(gcInfo, rvzInfo), nil
}

// RVZExtraInfo combines GCInfo and RVZInfo for the Extra field.
type RVZExtraInfo struct {
	GCInfo  *GCInfo
	RVZInfo *rvz.RVZInfo
}

// gcInfoToGameIdent converts GCInfo to GameIdent.
func gcInfoToGameIdent(info *GCInfo, rvzInfo *rvz.RVZInfo) *game.GameIdent {
	platform := game.PlatformGC
	if info.IsWii {
		platform = game.PlatformWii
	}

	// Build the full game ID (DiscID + GameCode + RegionCode)
	titleID := fmt.Sprintf("%c%s%c", info.DiscID, info.GameCode, info.RegionCode)

	version := info.Version
	discNumber := info.DiscNumber

	var extra any
	if rvzInfo != nil {
		extra = &RVZExtraInfo{
			GCInfo:  info,
			RVZInfo: rvzInfo,
		}
	} else {
		extra = info
	}

	return &game.GameIdent{
		Platform:   platform,
		TitleID:    titleID,
		Title:      info.Title,
		Regions:    []game.Region{decodeRegion(info.RegionCode)},
		MakerCode:  info.MakerCode,
		Version:    &version,
		DiscNumber: &discNumber,
		Extra:      extra,
	}
}

// decodeRegion converts a GameCube/Wii region code byte to a Region.
func decodeRegion(code byte) game.Region {
	switch code {
	case 'D':
		return game.RegionDE
	case 'E':
		return game.RegionUS
	case 'F':
		return game.RegionFR
	case 'I':
		return game.RegionIT
	case 'J':
		return game.RegionJP
	case 'K':
		return game.RegionKR
	case 'P':
		return game.RegionEU
	case 'R':
		return game.RegionUnknown // Russia - not in current Region constants
	case 'S':
		return game.RegionES
	case 'T':
		return game.RegionUnknown // Taiwan - not in current Region constants
	case 'U':
		return game.RegionAU
	default:
		return game.RegionUnknown
	}
}
