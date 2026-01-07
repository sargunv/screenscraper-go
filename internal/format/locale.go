package format

import (
	"github.com/Xuanwo/go-locale"
)

// GetPreferredLanguage returns the preferred base language code (e.g., "en", "fr", "de")
// based on system locale or override. Falls back to "en" if unavailable.
func GetPreferredLanguage(override string) string {
	if override != "" {
		return override
	}

	// Detect system locale using go-locale
	tag, err := locale.Detect()
	if err != nil {
		return "en"
	}

	// Extract base language (e.g., "en" from "en-US")
	base, _ := tag.Base()
	return base.String()
}
