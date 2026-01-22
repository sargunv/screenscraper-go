package region

import "strings"

// Hierarchy maps child regions to their parent regions
// Used for fallback when exact region match isn't available
var Hierarchy = map[string]string{
	"jp":  "asi",
	"kr":  "asi",
	"tw":  "asi",
	"cn":  "asi",
	"hk":  "asi",
	"de":  "eu",
	"fr":  "eu",
	"it":  "eu",
	"es":  "eu",
	"uk":  "eu",
	"nl":  "eu",
	"se":  "eu",
	"dk":  "eu",
	"fi":  "eu",
	"pt":  "eu",
	"us":  "ame",
	"ca":  "ame",
	"br":  "ame",
	"mex": "ame",
	"au":  "oce",
	"nz":  "oce",
	"asi": "wor",
	"eu":  "wor",
	"ame": "wor",
	"oce": "wor",
}

// ToLanguage maps region codes to language codes
// Used for selecting localized text
var ToLanguage = map[string]string{
	"jp":  "ja",
	"kr":  "ko",
	"tw":  "zh",
	"cn":  "zh",
	"hk":  "zh",
	"de":  "de",
	"fr":  "fr",
	"it":  "it",
	"es":  "es",
	"uk":  "en",
	"nl":  "nl",
	"se":  "sv",
	"dk":  "da",
	"fi":  "fi",
	"pt":  "pt",
	"us":  "en",
	"ca":  "en",
	"br":  "pt",
	"mex": "es",
	"au":  "en",
	"nz":  "en",
	"eu":  "en",
	"ame": "en",
	"wor": "en",
	"ss":  "en", // ScreenScraper default
}

// BuildSearchOrder creates an ordered list of regions to search
// based on ROM regions and user preferences, with fallback through hierarchy
func BuildSearchOrder(romRegions, userRegions []string) []string {
	seen := make(map[string]bool)
	var order []string

	addWithParents := func(region string) {
		for r := region; r != ""; {
			if !seen[r] {
				order = append(order, r)
				seen[r] = true
			}
			r = Hierarchy[r]
		}
	}

	// ROM regions first
	for _, r := range romRegions {
		addWithParents(r)
	}

	// User-specified regions
	for _, r := range userRegions {
		addWithParents(r)
	}

	// Ensure "wor" (world) is always included
	if !seen["wor"] {
		order = append(order, "wor")
		seen["wor"] = true
	}

	// Empty string for media with no region
	order = append(order, "")

	return order
}

// LocalizedEntry represents a text entry with a language
type LocalizedEntry struct {
	Language string
	Text     string
}

// SelectLocalizedText chooses the best text based on region preferences
func SelectLocalizedText(entries []LocalizedEntry, romRegions, userRegions []string) string {
	if len(entries) == 0 {
		return ""
	}

	// Build search order
	searchOrder := BuildSearchOrder(romRegions, userRegions)

	// Map to quickly check entries by language
	byLang := make(map[string]string)
	for _, e := range entries {
		if e.Text != "" {
			lang := strings.ToLower(e.Language)
			if _, exists := byLang[lang]; !exists {
				byLang[lang] = e.Text
			}
		}
	}

	// Search based on region -> language mapping
	for _, region := range searchOrder {
		if lang, ok := ToLanguage[region]; ok {
			if text, ok := byLang[lang]; ok {
				return text
			}
		}
	}

	// Fallback to English
	if text, ok := byLang["en"]; ok {
		return text
	}

	// Fallback to any available
	for _, e := range entries {
		if e.Text != "" {
			return e.Text
		}
	}

	return ""
}

// Media represents a media item with region information
type Media struct {
	Type   string
	Region string
	URL    string
	Format string
}

// SelectMedia finds the best media match for a given type and region preferences
func SelectMedia(available []Media, mediaType string, romRegions, userRegions []string) *Media {
	// Filter to matching media type
	var candidates []Media
	for _, m := range available {
		if m.Type == mediaType {
			candidates = append(candidates, m)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Build search order
	searchOrder := BuildSearchOrder(romRegions, userRegions)

	// Find first match in search order
	for _, region := range searchOrder {
		for i := range candidates {
			if candidates[i].Region == region {
				return &candidates[i]
			}
		}
	}

	// Fall back to any available
	return &candidates[0]
}
