package region

import (
	"regexp"
	"strings"
)

// filenamePatterns maps filename region tags to region codes
// Ordered by specificity (more specific patterns first)
var filenamePatterns = []struct {
	pattern string
	regions []string
}{
	// Multi-region patterns (more specific)
	{"(USA, Europe)", []string{"us", "eu"}},
	{"(Europe, USA)", []string{"eu", "us"}},
	{"(Japan, USA)", []string{"jp", "us"}},
	{"(USA, Japan)", []string{"us", "jp"}},
	{"(Japan, Europe)", []string{"jp", "eu"}},
	{"(Europe, Japan)", []string{"eu", "jp"}},
	{"(USA, Asia)", []string{"us", "asi"}},
	{"(Japan, Korea)", []string{"jp", "kr"}},
	{"(USA, Australia)", []string{"us", "au"}},
	{"(Europe, Australia)", []string{"eu", "au"}},

	// Single region patterns
	{"(USA)", []string{"us"}},
	{"(U)", []string{"us"}},
	{"(Japan)", []string{"jp"}},
	{"(J)", []string{"jp"}},
	{"(Europe)", []string{"eu"}},
	{"(E)", []string{"eu"}},
	{"(World)", []string{"wor"}},
	{"(W)", []string{"wor"}},
	{"(Germany)", []string{"de"}},
	{"(France)", []string{"fr"}},
	{"(Spain)", []string{"es"}},
	{"(Italy)", []string{"it"}},
	{"(Netherlands)", []string{"nl"}},
	{"(Sweden)", []string{"se"}},
	{"(Denmark)", []string{"dk"}},
	{"(Finland)", []string{"fi"}},
	{"(Portugal)", []string{"pt"}},
	{"(Korea)", []string{"kr"}},
	{"(China)", []string{"cn"}},
	{"(Taiwan)", []string{"tw"}},
	{"(Hong Kong)", []string{"hk"}},
	{"(Australia)", []string{"au"}},
	{"(Brazil)", []string{"br"}},
	{"(Canada)", []string{"ca"}},
	{"(Mexico)", []string{"mex"}},
	{"(Asia)", []string{"asi"}},
	{"(United Kingdom)", []string{"uk"}},
	{"(UK)", []string{"uk"}},

	// Rev patterns (still regional)
	{"(Rev A)", []string{}}, // No specific region, just revision
	{"(Rev B)", []string{}},
	{"(Rev 1)", []string{}},
	{"(Rev 2)", []string{}},
}

// languagePatterns map language tags to likely regions
// These appear in filenames like "(En,Fr,De)"
var languagePatterns = map[string]string{
	"En": "us",
	"Ja": "jp",
	"Fr": "fr",
	"De": "de",
	"Es": "es",
	"It": "it",
	"Nl": "nl",
	"Pt": "pt",
	"Sv": "se",
	"Da": "dk",
	"Fi": "fi",
	"Ko": "kr",
	"Zh": "cn",
}

// languageTagRegex matches language tags like "(En,Fr,De)" or "(En+Fr+De)"
var languageTagRegex = regexp.MustCompile(`\(([A-Z][a-z](?:[,+][A-Z][a-z])*)\)`)

// ParseFilename extracts region codes from a ROM filename
// Uses No-Intro and Redump naming conventions
func ParseFilename(filename string) []string {
	var regions []string
	seen := make(map[string]bool)

	// Check explicit region patterns first
	for _, p := range filenamePatterns {
		if strings.Contains(filename, p.pattern) {
			for _, r := range p.regions {
				if !seen[r] {
					regions = append(regions, r)
					seen[r] = true
				}
			}
		}
	}

	// If we found explicit regions, return them
	if len(regions) > 0 {
		return regions
	}

	// Try language tag patterns like "(En,Fr,De)"
	matches := languageTagRegex.FindAllStringSubmatch(filename, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// Split by comma or plus
			tags := strings.FieldsFunc(match[1], func(r rune) bool {
				return r == ',' || r == '+'
			})
			for _, tag := range tags {
				if region, ok := languagePatterns[tag]; ok {
					if !seen[region] {
						regions = append(regions, region)
						seen[region] = true
					}
				}
			}
		}
	}

	// If we found language-based regions, infer region hierarchy
	if len(regions) > 0 {
		// If multiple European languages, probably EU release
		euLangs := 0
		for _, r := range regions {
			switch r {
			case "fr", "de", "es", "it", "nl", "se", "dk", "fi", "pt":
				euLangs++
			}
		}
		if euLangs >= 2 && !seen["eu"] {
			regions = append(regions, "eu")
		}
		return regions
	}

	return nil
}

// Normalize converts various region representations to standard codes
func Normalize(region string) string {
	region = strings.ToLower(strings.TrimSpace(region))

	// Map common variations
	switch region {
	case "usa", "us", "u":
		return "us"
	case "japan", "jpn", "jp", "j":
		return "jp"
	case "europe", "eur", "eu", "e":
		return "eu"
	case "world", "wor", "w":
		return "wor"
	case "germany", "ger", "deu":
		return "de"
	case "france", "fra":
		return "fr"
	case "spain", "spa":
		return "es"
	case "italy", "ita":
		return "it"
	case "korea", "kor":
		return "kr"
	case "china", "chn":
		return "cn"
	case "taiwan":
		return "tw"
	case "australia", "aus":
		return "au"
	case "brazil", "bra":
		return "br"
	case "asia", "asi":
		return "asi"
	case "uk", "united kingdom", "gb", "gbr":
		return "uk"
	default:
		return region
	}
}
