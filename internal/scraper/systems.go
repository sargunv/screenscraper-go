package scraper

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sargunv/rom-tools/lib/core"
)

// SystemMapping maps platform names to Screenscraper system IDs.
// Platform names can be romident Platform values, recalbox names, or common aliases.
var SystemMapping = map[string]string{
	// Nintendo consoles (from screenscraper list systems)
	"nes":          "3",
	"famicom":      "3", // romident Platform
	"snes":         "4",
	"superfamicom": "4", // romident Platform
	"n64":          "14",
	"nintendo64":   "14", // romident Platform
	"gc":           "13",
	"gamecube":     "13", // romident Platform
	"ngc":          "13", // alias
	"wii":          "16",
	"wiiu":         "18",
	"fds":          "106", // Famicom Disk System

	// Nintendo handhelds
	"gb":             "9",
	"gameboy":        "9", // romident Platform
	"gbc":            "10",
	"gameboycolor":   "10", // romident Platform
	"gba":            "12",
	"gameboyadvance": "12", // romident Platform
	"nds":            "15",
	"ds":             "15", // romident Platform
	"dsi":            "15", // romident Platform (same system ID)
	"3ds":            "17",
	"virtualboy":     "11",
	"vb":             "11", // alias

	// Sega consoles
	"megadrive":    "1",
	"md":           "1", // alias
	"genesis":      "1", // alias
	"mastersystem": "2",
	"sms":          "2", // alias
	"sega32x":      "19",
	"32x":          "19", // alias
	"segacd":       "20",
	"megacd":       "20", // alias
	"gamegear":     "21",
	"gg":           "21", // alias
	"saturn":       "22",
	"dreamcast":    "23",
	"dc":           "23", // alias

	// Sony consoles
	"psx":          "57",
	"ps1":          "57", // alias
	"playstation":  "57", // romident Platform
	"ps2":          "58",
	"playstation2": "58", // romident Platform
	"ps3":          "59",
	"playstation3": "59", // romident Platform

	// Sony handhelds
	"psp":    "61",
	"psvita": "62",
	"vita":   "62", // alias

	// Microsoft
	"xbox":    "32",
	"xbox360": "33",
	"xboxone": "34",

	// NEC
	"pcengine":     "31",
	"pce":          "31", // alias
	"turbografx16": "31", // alias
	"tg16":         "31", // alias
	"supergrafx":   "105",
	"sgx":          "105", // alias
	"pcfx":         "72",

	// SNK
	"neogeo":       "142",
	"ng":           "142", // alias
	"neogeocd":     "70",
	"ngcd":         "70", // alias
	"ngp":          "25",
	"neogeopocket": "25", // alias
	"ngpc":         "82",

	// Atari
	"atari2600":   "26",
	"2600":        "26", // alias
	"atari5200":   "40",
	"5200":        "40", // alias
	"atari7800":   "41",
	"7800":        "41", // alias
	"lynx":        "28",
	"atarilynx":   "28", // alias
	"jaguar":      "27",
	"atarijaguar": "27", // alias

	// Bandai
	"wonderswan":      "45",
	"ws":              "45", // alias
	"wonderswancolor": "46",
	"wsc":             "46", // alias

	// Other
	"colecovision":  "48",
	"intellivision": "115",
	"vectrex":       "102",
	"3do":           "29",
}

// LookupSystemID converts a platform name to a Screenscraper system ID.
// Accepts romident Platform values, recalbox names, or common aliases.
// Returns error if the platform is not recognized.
func LookupSystemID(platform string) (string, error) {
	// Normalize input
	normalized := strings.ToLower(strings.TrimSpace(platform))

	// Look up in mapping
	if id, ok := SystemMapping[normalized]; ok {
		return id, nil
	}

	return "", fmt.Errorf("unknown system: %q (use 'rom-tools screenscraper list systems' to see all systems)", platform)
}

// AvailableSystems returns a sorted list of commonly used system names.
func AvailableSystems() []string {
	// Collect unique primary names (prefer short names)
	seen := make(map[string]bool)
	result := make([]string, 0)

	// Primary short names (recalbox style)
	primaryNames := []string{
		// Nintendo
		"nes", "snes", "n64", "gc", "wii", "wiiu", "fds",
		"gb", "gbc", "gba", "nds", "3ds", "virtualboy",
		// Sega
		"megadrive", "mastersystem", "sega32x", "segacd", "gamegear", "saturn", "dreamcast",
		// Sony
		"psx", "ps2", "ps3", "psp", "psvita",
		// Microsoft
		"xbox", "xbox360",
		// NEC
		"pcengine", "supergrafx", "pcfx",
		// SNK
		"neogeo", "neogeocd", "ngp", "ngpc",
		// Atari
		"atari2600", "atari5200", "atari7800", "lynx", "jaguar",
		// Bandai
		"wonderswan", "wonderswancolor",
		// Other
		"colecovision", "vectrex", "3do",
	}

	for _, name := range primaryNames {
		if _, ok := SystemMapping[name]; ok && !seen[name] {
			result = append(result, name)
			seen[name] = true
		}
	}

	// Also include romident Platform values that aren't already covered
	romidentPlatforms := []core.Platform{
		core.PlatformNES, core.PlatformSNES, core.PlatformN64, core.PlatformGC,
		core.PlatformWii, core.PlatformWiiU, core.PlatformGB, core.PlatformGBC,
		core.PlatformGBA, core.PlatformNDS, core.PlatformDSi, core.Platform3DS,
		core.PlatformPS1, core.PlatformPS2, core.PlatformPS3, core.PlatformPSP,
		core.PlatformPSVita, core.PlatformMS, core.PlatformMD, core.PlatformSaturn,
		core.PlatformDreamcast, core.PlatformGameGear, core.PlatformXbox, core.PlatformXbox360,
	}

	for _, p := range romidentPlatforms {
		name := string(p)
		if _, ok := SystemMapping[name]; ok && !seen[name] {
			result = append(result, name)
			seen[name] = true
		}
	}

	sort.Strings(result)
	return result
}
