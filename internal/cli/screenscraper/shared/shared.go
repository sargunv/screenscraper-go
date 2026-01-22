package shared

import (
	"fmt"
	"os"

	"github.com/sargunv/rom-tools/lib/screenscraper"
)

// Shared state for screenscraper CLI commands
var (
	JsonOutput bool
	Locale     string
	Client     *screenscraper.ScreenscraperClient
)

// NewClientFromEnv creates a client from environment variables
// Returns error if dev credentials are missing
func NewClientFromEnv(appName string) (*screenscraper.ScreenscraperClient, error) {
	devID := os.Getenv("SCREENSCRAPER_DEV_USER")
	devPassword := os.Getenv("SCREENSCRAPER_DEV_PASSWORD")
	ssID := os.Getenv("SCREENSCRAPER_ID")
	ssPassword := os.Getenv("SCREENSCRAPER_PASSWORD")

	if devID == "" || devPassword == "" {
		return nil, fmt.Errorf("screenscraper credentials required: set SCREENSCRAPER_DEV_USER and SCREENSCRAPER_DEV_PASSWORD")
	}

	return screenscraper.NewScreenscraperClient(devID, devPassword, appName, ssID, ssPassword)
}
