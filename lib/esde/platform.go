package esde

import "github.com/sargunv/rom-tools/lib/core"

// PlatformDirectory returns the ES-DE ROM directory name for a platform.
// These match the default ES-DE system directories.
func PlatformDirectory(platform core.Platform) string {
	dir, ok := platformDirs[platform]
	if !ok {
		return string(platform) // Fallback to platform name
	}
	return dir
}

// platformDirs maps Platform values to ES-DE directory names.
// These are the default directories ES-DE uses.
var platformDirs = map[core.Platform]string{
	// Nintendo consoles
	core.PlatformNES:    "nes",
	core.PlatformSNES:   "snes",
	core.PlatformN64:    "n64",
	core.PlatformGC:     "gc",
	core.PlatformWii:    "wii",
	core.PlatformWiiU:   "wiiu",
	core.PlatformSwitch: "switch",

	// Nintendo handhelds
	core.PlatformGB:     "gb",
	core.PlatformGBC:    "gbc",
	core.PlatformGBA:    "gba",
	core.PlatformNDS:    "nds",
	core.PlatformDSi:    "nds", // DSi uses same directory as DS
	core.Platform3DS:    "3ds",
	core.PlatformNew3DS: "3ds", // New 3DS uses same directory as 3DS

	// Sony consoles
	core.PlatformPS1: "psx",
	core.PlatformPS2: "ps2",
	core.PlatformPS3: "ps3",

	// Sony handhelds
	core.PlatformPSP:    "psp",
	core.PlatformPSVita: "psvita",

	// Sega consoles
	core.PlatformMS:        "mastersystem",
	core.PlatformMD:        "megadrive",
	core.PlatformSaturn:    "saturn",
	core.PlatformDreamcast: "dreamcast",

	// Sega handhelds
	core.PlatformGameGear: "gamegear",

	// Microsoft
	core.PlatformXbox:    "xbox",
	core.PlatformXbox360: "xbox360",
}

// MediaTypes defines standard ES-DE media type directories.
type MediaType string

const (
	MediaTypeScreenshot  MediaType = "screenshots"
	MediaTypeTitlescreen MediaType = "titlescreens"
	MediaTypeBoxFront    MediaType = "covers"
	MediaTypeBoxBack     MediaType = "backcovers"
	MediaTypeWheel       MediaType = "wheels"
	MediaTypeMarquee     MediaType = "marquees"
	MediaTypeFanart      MediaType = "fanart"
	MediaTypeVideo       MediaType = "videos"
	MediaType3DBox       MediaType = "3dboxes"
	MediaTypeMix         MediaType = "miximages"
)

// MediaPath returns the relative path for a media file.
// For example: MediaPath(MediaTypeScreenshot, "Super Mario Bros") returns "screenshots/Super Mario Bros.png"
func MediaPath(mediaType MediaType, baseName, extension string) string {
	return string(mediaType) + "/" + baseName + extension
}
