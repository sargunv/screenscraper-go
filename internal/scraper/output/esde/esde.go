package esde

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sargunv/rom-tools/internal/region"
	"github.com/sargunv/rom-tools/internal/scraper"
	"github.com/sargunv/rom-tools/lib/esde"
	"github.com/sargunv/rom-tools/lib/screenscraper"
)

// Generator generates ES-DE compatible output
type Generator struct {
	gamelistPath string
	mediaDir     string
	overwrite    bool
	regions      []string
}

// NewGenerator creates a new ES-DE output generator
func NewGenerator(gamelistPath, mediaDir string, overwrite bool, preferredRegions []string) *Generator {
	return &Generator{
		gamelistPath: gamelistPath,
		mediaDir:     mediaDir,
		overwrite:    overwrite,
		regions:      preferredRegions,
	}
}

// Generate creates ES-DE output from scrape results
func (g *Generator) Generate(results *scraper.ScrapeResults) error {
	// Load existing gamelist if present
	var existing *esde.GameList
	if data, err := os.ReadFile(g.gamelistPath); err == nil {
		existing = &esde.GameList{}
		if err := xml.Unmarshal(data, existing); err != nil {
			existing = nil // Ignore parse errors, start fresh
		}
	}

	// Convert results to games
	newGames := make([]esde.Game, 0)
	for _, result := range results.Results {
		if result.Game == nil {
			continue // Skip not found or errored
		}

		game := g.resultToGame(result)
		newGames = append(newGames, game)

		// Save media files
		if err := g.saveMedia(result); err != nil {
			// Log but don't fail
			continue
		}
	}

	// Merge with existing
	finalList := g.merge(existing, &esde.GameList{Games: newGames})

	// Write gamelist
	return g.writeGameList(finalList)
}

// resultToGame converts a scrape result to an ES-DE game entry
func (g *Generator) resultToGame(result *scraper.ScrapeResult) esde.Game {
	entry := result.Entry
	ssGame := result.Game

	// Get localized text
	romRegions := entry.Regions
	userRegions := g.regions

	// Build name
	name := selectName(ssGame.Names, romRegions, userRegions)
	if name == "" {
		name = ssGame.Name
	}
	if name == "" {
		name = entry.Name
	}

	// Build description
	desc := selectLocalizedText(ssGame.Synopsis, romRegions, userRegions)

	// Build genre
	var genres []string
	for _, genre := range ssGame.Genres {
		genreName := selectLocalizedText(genre.Names, romRegions, userRegions)
		if genreName != "" {
			genres = append(genres, genreName)
		}
	}

	// Format rating (0-1 scale)
	var rating float64
	if ssGame.Note.Text != "" {
		// Screenscraper uses 0-20 scale, convert to 0-1
		if noteVal, err := strconv.ParseFloat(ssGame.Note.Text, 64); err == nil {
			rating = noteVal / 20.0
		}
	}

	// Format release date (YYYYMMDDTHHMMSS)
	releaseDate := selectReleaseDate(ssGame.Dates, romRegions, userRegions)

	// Get developer, publisher
	developer := ssGame.Developer.Text
	publisher := ssGame.Publisher.Text

	// Parse players
	var players int
	if ssGame.Players.Text != "" {
		if p, err := strconv.Atoi(ssGame.Players.Text); err == nil {
			players = p
		}
	}

	return esde.Game{
		Path:        "./" + entry.BaseName + filepath.Ext(entry.Name),
		Name:        name,
		Desc:        desc,
		Rating:      rating,
		ReleaseDate: releaseDate,
		Developer:   developer,
		Publisher:   publisher,
		Genre:       strings.Join(genres, ", "),
		Players:     players,
	}
}

// saveMedia saves media files for a result
func (g *Generator) saveMedia(result *scraper.ScrapeResult) error {
	for _, relativePath := range result.Media {
		// The relativePath is already like "screenshots/GameName.png"
		fullPath := filepath.Join(g.mediaDir, relativePath)

		// Check if exists and skip if not overwriting
		if !g.overwrite {
			if _, err := os.Stat(fullPath); err == nil {
				continue // File exists, skip
			}
		}

		// Create directory
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create media directory %s: %w", dir, err)
		}
	}

	return nil
}

// merge combines existing and new gamelists
func (g *Generator) merge(existing, new *esde.GameList) *esde.GameList {
	if existing == nil {
		return new
	}

	// Index existing by path
	existingByPath := make(map[string]*esde.Game)
	for i := range existing.Games {
		existingByPath[existing.Games[i].Path] = &existing.Games[i]
	}

	result := &esde.GameList{
		Games: make([]esde.Game, 0, len(existing.Games)+len(new.Games)),
	}

	// Process new games
	for _, game := range new.Games {
		if existingGame, ok := existingByPath[game.Path]; ok {
			if g.overwrite {
				result.Games = append(result.Games, game) // Use new
			} else {
				result.Games = append(result.Games, *existingGame) // Keep existing
			}
			delete(existingByPath, game.Path)
		} else {
			result.Games = append(result.Games, game) // Add new
		}
	}

	// Add remaining existing games
	for _, game := range existingByPath {
		result.Games = append(result.Games, *game)
	}

	return result
}

// writeGameList writes the gamelist.xml file
func (g *Generator) writeGameList(list *esde.GameList) error {
	// Ensure directory exists
	dir := filepath.Dir(g.gamelistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create gamelist directory: %w", err)
	}

	// Marshal with indentation
	data, err := xml.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal gamelist: %w", err)
	}

	// Add XML header
	output := []byte(xml.Header + string(data) + "\n")

	// Write file
	if err := os.WriteFile(g.gamelistPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write gamelist: %w", err)
	}

	return nil
}

// Helper functions for localized content selection

func selectName(names []screenscraper.NameEntry, romRegions, userRegions []string) string {
	if len(names) == 0 {
		return ""
	}

	searchOrder := region.BuildSearchOrder(romRegions, userRegions)

	for _, r := range searchOrder {
		for _, n := range names {
			if n.Region == r && n.Text != "" {
				return n.Text
			}
		}
	}

	// Fallback to any
	for _, n := range names {
		if n.Text != "" {
			return n.Text
		}
	}

	return ""
}

func selectLocalizedText(entries []screenscraper.LocalizedName, romRegions, userRegions []string) string {
	if len(entries) == 0 {
		return ""
	}

	// Convert to region.LocalizedEntry
	regionEntries := make([]region.LocalizedEntry, len(entries))
	for i, e := range entries {
		regionEntries[i] = region.LocalizedEntry{
			Language: e.Language,
			Text:     e.Text,
		}
	}

	return region.SelectLocalizedText(regionEntries, romRegions, userRegions)
}

func selectReleaseDate(dates []screenscraper.DateEntry, romRegions, userRegions []string) esde.DateTime {
	if len(dates) == 0 {
		return esde.DateTime{}
	}

	searchOrder := region.BuildSearchOrder(romRegions, userRegions)

	for _, r := range searchOrder {
		for _, d := range dates {
			if d.Region == r && d.Text != "" {
				return parseDate(d.Text)
			}
		}
	}

	// Fallback to any
	for _, d := range dates {
		if d.Text != "" {
			return parseDate(d.Text)
		}
	}

	return esde.DateTime{}
}

// parseDate converts screenscraper date format to esde.DateTime
// Screenscraper: "1991-06-23" or "1991"
func parseDate(date string) esde.DateTime {
	// Remove any dashes
	clean := strings.ReplaceAll(date, "-", "")

	// Pad with zeros if needed
	var dateStr string
	switch len(clean) {
	case 4: // Just year
		dateStr = clean + "0101T000000"
	case 6: // Year and month
		dateStr = clean + "01T000000"
	case 8: // Full date
		dateStr = clean + "T000000"
	default:
		dateStr = clean
	}

	t, err := time.Parse(esde.DateTimeFormat, dateStr)
	if err != nil {
		return esde.DateTime{}
	}
	return esde.DateTime{Time: t}
}
