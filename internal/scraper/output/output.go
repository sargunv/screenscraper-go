package output

import (
	"github.com/sargunv/rom-tools/internal/scraper"
)

// Generator is the interface for output generators
type Generator interface {
	// Generate creates output from scrape results
	Generate(results *scraper.ScrapeResults) error
}
