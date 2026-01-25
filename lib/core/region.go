package core

// Region represents a geographic region for ROM/asset matching.
// Regions form a hierarchy (e.g., Germany -> Europe -> World) used for
// fallback matching when exact region assets aren't available.
type Region string

const (
	RegionUnknown Region = ""

	// Top level regions (parent is World, or no parent for World itself)
	RegionWorld      Region = "World"
	RegionEurope     Region = "Europe"
	RegionAsia       Region = "Asia"
	RegionAmericas   Region = "Americas"
	RegionOceania    Region = "Oceania"
	RegionMiddleEast Region = "Middle East"
	RegionAfrica     Region = "Africa"

	// Europe children
	RegionGermany     Region = "Germany"
	RegionFrance      Region = "France"
	RegionUK          Region = "UK"
	RegionSpain       Region = "Spain"
	RegionItaly       Region = "Italy"
	RegionNetherlands Region = "Netherlands"
	RegionSweden      Region = "Sweden"
	RegionDenmark     Region = "Denmark"
	RegionFinland     Region = "Finland"
	RegionNorway      Region = "Norway"
	RegionPortugal    Region = "Portugal"
	RegionPoland      Region = "Poland"
	RegionCzechia     Region = "Czechia"
	RegionHungary     Region = "Hungary"
	RegionSlovakia    Region = "Slovakia"
	RegionBulgaria    Region = "Bulgaria"
	RegionGreece      Region = "Greece"
	RegionRussia      Region = "Russia"

	// Asia children
	RegionJapan  Region = "Japan"
	RegionChina  Region = "China"
	RegionKorea  Region = "Korea"
	RegionTaiwan Region = "Taiwan"

	// America children
	RegionUSA    Region = "USA"
	RegionCanada Region = "Canada"
	RegionBrazil Region = "Brazil"
	RegionMexico Region = "Mexico"
	RegionChile  Region = "Chile"
	RegionPeru   Region = "Peru"

	// Oceania children
	RegionAustralia  Region = "Australia"
	RegionNewZealand Region = "New Zealand"

	// Middle East children
	RegionIsrael Region = "Israel"
	RegionTurkey Region = "Turkey"
	RegionKuwait Region = "Kuwait"
	RegionUAE    Region = "UAE"

	// Africa children
	RegionSouthAfrica Region = "South Africa"
)

// regionParents maps each region to its parent in the hierarchy.
var regionParents = map[Region]Region{
	// Continental regions -> World
	RegionEurope:     RegionWorld,
	RegionAsia:       RegionWorld,
	RegionAmericas:   RegionWorld,
	RegionOceania:    RegionWorld,
	RegionMiddleEast: RegionWorld,
	RegionAfrica:     RegionWorld,

	// Europe children
	RegionGermany:     RegionEurope,
	RegionFrance:      RegionEurope,
	RegionUK:          RegionEurope,
	RegionSpain:       RegionEurope,
	RegionItaly:       RegionEurope,
	RegionNetherlands: RegionEurope,
	RegionSweden:      RegionEurope,
	RegionDenmark:     RegionEurope,
	RegionFinland:     RegionEurope,
	RegionNorway:      RegionEurope,
	RegionPortugal:    RegionEurope,
	RegionPoland:      RegionEurope,
	RegionCzechia:     RegionEurope,
	RegionHungary:     RegionEurope,
	RegionSlovakia:    RegionEurope,
	RegionBulgaria:    RegionEurope,
	RegionGreece:      RegionEurope,
	RegionRussia:      RegionEurope,

	// Asia children
	RegionJapan:  RegionAsia,
	RegionChina:  RegionAsia,
	RegionKorea:  RegionAsia,
	RegionTaiwan: RegionAsia,

	// America children
	RegionUSA:    RegionAmericas,
	RegionCanada: RegionAmericas,
	RegionBrazil: RegionAmericas,
	RegionMexico: RegionAmericas,
	RegionChile:  RegionAmericas,
	RegionPeru:   RegionAmericas,

	// Oceania children
	RegionAustralia:  RegionOceania,
	RegionNewZealand: RegionOceania,

	// Middle East children
	RegionIsrael: RegionMiddleEast,
	RegionTurkey: RegionMiddleEast,
	RegionKuwait: RegionMiddleEast,
	RegionUAE:    RegionMiddleEast,

	// Africa children
	RegionSouthAfrica: RegionAfrica,
}

// Parent returns this region's parent in the hierarchy.
// Returns RegionWorld for top-level continental regions.
// Returns RegionUnknown for RegionWorld and RegionUnknown.
func (r Region) Parent() Region {
	if parent, ok := regionParents[r]; ok {
		return parent
	}
	return RegionUnknown
}

// Ancestors returns the chain of ancestors from this region up to World.
// For example, RegionGermany.Ancestors() returns [RegionEurope, RegionWorld].
// Returns nil for RegionWorld and RegionUnknown.
func (r Region) Ancestors() []Region {
	var ancestors []Region
	for p := r.Parent(); p != RegionUnknown; p = p.Parent() {
		ancestors = append(ancestors, p)
	}
	return ancestors
}

// IsAncestorOf returns true if r is an ancestor of other in the hierarchy,
// along with the distance (number of hops from other to r).
// For example, RegionEurope.IsAncestorOf(RegionGermany) returns (true, 1).
// Returns (false, -1) if r is not an ancestor of other.
func (r Region) IsAncestorOf(other Region) (bool, int) {
	dist := 0
	for p := other.Parent(); p != RegionUnknown; p = p.Parent() {
		dist++
		if p == r {
			return true, dist
		}
	}
	return false, -1
}

// IsDescendantOf returns true if r is a descendant of other in the hierarchy,
// along with the distance (number of hops from r to other).
// For example, RegionGermany.IsDescendantOf(RegionEurope) returns (true, 1).
// Returns (false, -1) if r is not a descendant of other.
func (r Region) IsDescendantOf(other Region) (bool, int) {
	return other.IsAncestorOf(r)
}
