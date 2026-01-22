package scrape

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sargunv/rom-tools/internal/cache"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/scraper"
	"github.com/sargunv/rom-tools/internal/scraper/output/esde"
	"github.com/sargunv/rom-tools/lib/datfile"
)

var (
	// Input
	datPath    string
	inputPath  string
	systemName string

	// Output - ES-DE
	esdeGamelist string
	esdeMedia    string

	// Media
	mediaTypes []string

	// Regions
	regions []string

	// Hash mode
	fastMode bool
	slowMode bool

	// Cache
	cacheAge  time.Duration
	noCache   bool
	cacheOnly bool

	// Overwrite
	overwrite bool

	// Network
	httpTimeout  time.Duration
	threadsLimit int

	// Filter
	filterExpr string

	// Other
	dryRun     bool
	jsonOutput bool
)

var Cmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape metadata for ROM collections",
	Long: `Batch scrape metadata and media for ROM files.

Scans the input (DAT file or ROM directory), identifies games using hashes,
fetches metadata from Screenscraper, downloads media files, and generates
output in the specified format(s).

Example:
  # Scrape from DAT file to ES-DE format
  rom-tools scrape --system megadrive --dat megadrive.dat \
      --esde-gamelist ./roms/megadrive/gamelist.xml \
      --esde-media ./roms/megadrive/media

  # Scrape with custom media types and regions
  rom-tools scrape --system gba --dat gba.dat \
      --esde-gamelist ./gba/gamelist.xml \
      --esde-media ./gba/media \
      --media screenshots,covers,3dboxes,marquees,videos \
      --regions jp,us,eu

  # Dry run to see what would be scraped
  rom-tools scrape --system snes --dat snes.dat --dry-run

Common systems: megadrive, gba, snes, nes, psx, ps2, dreamcast, n64, nds, gb, gbc.
Use 'rom-tools screenscraper list systems' to see all available systems.`,
	RunE: runScrape,
}

func init() {
	// Input flags
	Cmd.Flags().StringVarP(&datPath, "dat", "d", "", "Path to DAT file (Logiqx XML format)")
	Cmd.Flags().StringVarP(&inputPath, "input", "i", "", "Path to ROM directory (not yet implemented)")
	Cmd.Flags().StringVarP(&systemName, "system", "s", "", "System name or ID (e.g., megadrive, gba, snes, psx)")
	Cmd.MarkFlagRequired("system")

	// Output flags - ES-DE
	Cmd.Flags().StringVar(&esdeGamelist, "esde-gamelist", "", "Path for ES-DE gamelist.xml")
	Cmd.Flags().StringVar(&esdeMedia, "esde-media", "", "Path for ES-DE media folder")

	// Media flags
	Cmd.Flags().StringSliceVarP(&mediaTypes, "media", "m", scraper.DefaultMediaTypes(),
		"Media types to download: screenshots,titlescreens,covers,3dboxes,marquees,fanart,videos,physicalmedia,backcovers")

	// Region flags
	Cmd.Flags().StringSliceVarP(&regions, "regions", "r", []string{"us", "eu", "jp"},
		"Preferred regions in order")

	// Hash mode flags
	Cmd.Flags().BoolVar(&fastMode, "fast", false, "Skip hash calculation for large files")
	Cmd.Flags().BoolVar(&slowMode, "slow", false, "Calculate full hashes for archives")
	Cmd.MarkFlagsMutuallyExclusive("fast", "slow")

	// Cache flags
	Cmd.Flags().DurationVar(&cacheAge, "cache-age", 720*time.Hour, "Maximum cache age (default 30 days)")
	Cmd.Flags().BoolVar(&noCache, "no-cache", false, "Don't read from cache (still writes to cache)")
	Cmd.Flags().BoolVar(&cacheOnly, "cache-only", false, "Only use cached data, no API calls")
	Cmd.MarkFlagsMutuallyExclusive("no-cache", "cache-only")

	// Overwrite flag
	Cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing media files and gamelist entries")

	// Network flags
	Cmd.Flags().DurationVar(&httpTimeout, "http-timeout", 5*time.Minute, "HTTP request timeout (e.g., 30s, 2m, 5m)")
	Cmd.Flags().IntVar(&threadsLimit, "threads", 0, "Max concurrent API requests (0 = use account limit)")

	// Filter flags
	Cmd.Flags().StringVar(&filterExpr, "filter", "true", "Filter expression for which games to scrape (e.g., 'missing.metadata', 'missing.covers or missing.videos')")

	// Other flags
	Cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Parse input and show what would be scraped")
	Cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output final results as JSON")
}

func runScrape(cmd *cobra.Command, args []string) error {
	// Resolve system name to ID
	systemID, err := scraper.LookupSystemID(systemName)
	if err != nil {
		return err
	}

	// Validate input
	if datPath == "" && inputPath == "" {
		return fmt.Errorf("either --dat or --input is required")
	}
	if datPath != "" && inputPath != "" {
		return fmt.Errorf("cannot specify both --dat and --input")
	}
	if inputPath != "" {
		return fmt.Errorf("--input (ROM directory) is not yet implemented, use --dat")
	}

	// Validate filter expression early (before dry-run or output validation)
	filter, err := scraper.NewFilter(filterExpr)
	if err != nil {
		return err
	}

	// Build filter config for checking what's missing
	gamelistPath := normalizeGamelistPath(esdeGamelist)
	filterConfig := scraper.NewFilterConfig(gamelistPath, esdeMedia)

	// Dry run mode (doesn't require output targets)
	if dryRun {
		return runDryRun(filter, filterConfig)
	}

	// Validate output
	if esdeGamelist == "" && esdeMedia == "" {
		return fmt.Errorf("at least one output target is required (--esde-gamelist, --esde-media)")
	}

	// Normalize gamelist path
	esdeGamelist = normalizeGamelistPath(esdeGamelist)

	// Initialize client from environment variables
	client, err := shared.NewClientFromEnv("rom-tools")
	if err != nil {
		return err
	}

	// Validation complete - don't show help for errors from here on
	cmd.SilenceUsage = true

	// Note: HTTP timeout is configured at client creation time in the library

	// Set up signal handling early so we can cancel during setup if needed
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get user info for rate limits
	fmt.Print("Connecting to Screenscraper...")
	userInfoResp, err := client.GetUserInfoWithResponse(ctx)
	fmt.Print("\r\033[K") // Clear the line
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	if userInfoResp.JSON200 == nil || userInfoResp.JSON200.Response.User.Id == "" {
		return fmt.Errorf("failed to get user info: invalid response")
	}

	userInfo := userInfoResp.JSON200.Response.User
	maxThreadsStr := userInfo.MaxThreads
	maxReqPerMinStr := userInfo.MaxRequestsPerMin
	maxThreads, _ := strconv.Atoi(maxThreadsStr)
	maxReqPerMin, _ := strconv.Atoi(maxReqPerMinStr)

	if maxThreads == 0 {
		maxThreads = 1
	}
	if maxReqPerMin == 0 {
		maxReqPerMin = 60
	}

	// Apply user-specified thread limit
	if threadsLimit > 0 && threadsLimit < maxThreads {
		maxThreads = threadsLimit
	}

	// Initialize cache
	cacheDir, err := cache.DefaultCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	cacheMode := cache.ModeNormal
	if noCache {
		cacheMode = cache.ModeNoRead
	} else if cacheOnly {
		cacheMode = cache.ModeReadOnly
	}

	diskCache, err := cache.New(cacheDir, cacheAge, cacheMode)
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Build config
	config := &scraper.Config{
		SystemID:          systemID,
		MediaTypes:        mediaTypes,
		PreferredRegions:  regions,
		MediaOutputDir:    esdeMedia,
		SkipCacheRead:     noCache,
		SkipCacheWrite:    cacheOnly,
		Overwrite:         overwrite,
		MaxThreads:        maxThreads,
		MaxRequestsPerMin: maxReqPerMin,
		Filter:            filter,
		FilterConfig:      filterConfig,
	}

	// Create scraper
	s := scraper.New(client, diskCache, config)

	// Parse DAT to get total count for progress
	dat, err := datfile.Parse(datPath)
	if err != nil {
		return fmt.Errorf("failed to parse DAT file: %w", err)
	}

	// Count non-BIOS entries and apply filter to get actual scrape count
	totalInDat := 0
	toScrape := 0
	for _, game := range dat.Games {
		if isBIOS(game) || len(game.ROMs) == 0 {
			continue
		}
		totalInDat++

		// Apply filter to count how many will actually be scraped
		rom := game.ROMs[0]
		baseName := scraper.BaseName(rom.Name)
		ctx := scraper.BuildFilterContext(baseName, filterConfig)
		if shouldScrape, err := filter.ShouldScrape(ctx); err == nil && shouldScrape {
			toScrape++
		}
	}

	fmt.Printf("Found %d games in DAT file (excluding BIOS)\n", totalInDat)
	if filterExpr != "true" {
		fmt.Printf("Filter: %s\n", filterExpr)
		fmt.Printf("To scrape: %d (filtered out: %d)\n", toScrape, totalInDat-toScrape)
	}
	fmt.Printf("Using %d threads, %d req/min\n\n", maxThreads, maxReqPerMin)

	// Use filtered count for progress tracking
	total := toScrape

	// Run with TUI if terminal, otherwise simple output
	var results *scraper.ScrapeResults

	if !jsonOutput && isTerminal() {
		// Create and run TUI
		model := scraper.NewModel(total, maxThreads, len(mediaTypes), s.Updates(), s.RateLimiterStats)

		// Run scraper in background
		resultsChan := make(chan *scraper.ScrapeResults, 1)
		go func() {
			res, _ := s.ScrapeFromDAT(ctx, datPath)
			resultsChan <- res
		}()

		// Run TUI with context so it exits on Ctrl+C
		p := tea.NewProgram(model, tea.WithContext(ctx))
		if _, err := p.Run(); err != nil && ctx.Err() == nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		// Cancel context to stop scraper if TUI exited early (user pressed 'q')
		cancel()

		// Wait for scraper to finish (should be quick after cancellation)
		results = <-resultsChan
	} else {
		// Simple output mode
		results, err = s.ScrapeFromDAT(ctx, datPath)
		if err != nil {
			return fmt.Errorf("scrape failed: %w", err)
		}
	}

	cancelled := ctx.Err() != nil

	// Generate output (even if cancelled, save partial results)
	if results != nil && (esdeGamelist != "" || esdeMedia != "") {
		mediaDir := esdeMedia
		if mediaDir == "" && esdeGamelist != "" {
			// Default to gamelist directory + media
			mediaDir = esdeGamelist[:len(esdeGamelist)-len("/gamelist.xml")] + "/media"
		}

		generator := esde.NewGenerator(esdeGamelist, mediaDir, overwrite, regions)
		if err := generator.Generate(results); err != nil {
			return fmt.Errorf("failed to generate ES-DE output: %w", err)
		}
	}

	// Get final stats
	stats := s.RateLimiterStats()

	// Output results
	if jsonOutput {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"cancelled":        cancelled,
			"total":            results.TotalEntries,
			"filtered_out":     results.FilteredOut,
			"found":            results.Found,
			"not_found":        results.NotFound,
			"skipped":          results.Skipped,
			"errors":           results.Errors,
			"media_downloaded": results.MediaDownloaded,
			"cache_hits":       results.CacheHits,
			"api_calls":        stats.TotalRequests,
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		// Progress summary
		processed := results.Found + results.NotFound + results.Skipped + results.Errors
		pct := 0.0
		if results.TotalEntries > 0 {
			pct = float64(processed) / float64(results.TotalEntries) * 100
		}
		fmt.Printf("\n")
		fmt.Printf(" Progress: %d/%d (%.0f%%)", processed, results.TotalEntries, pct)
		if cancelled {
			fmt.Printf(" [cancelled]")
		}
		if results.FilteredOut > 0 {
			fmt.Printf(" [%d filtered out]", results.FilteredOut)
		}
		fmt.Printf("\n\n")

		// Stats (matching TUI format)
		fmt.Printf(" Found: %d    Not Found: %d    Skipped: %d    Errors: %d\n",
			results.Found, results.NotFound, results.Skipped, results.Errors)
		fmt.Printf(" Media: %d downloaded    Cache hits: %d\n\n",
			results.MediaDownloaded, results.CacheHits)

		// API stats
		fmt.Printf(" API: %d calls completed\n", stats.TotalRequests)
	}

	return nil
}

func runDryRun(filter *scraper.Filter, filterConfig *scraper.FilterConfig) error {
	dat, err := datfile.Parse(datPath)
	if err != nil {
		return fmt.Errorf("failed to parse DAT file: %w", err)
	}

	stats := getStats(dat)

	// Count how many games pass the filter
	nonBIOS := 0
	toScrape := 0
	for _, game := range dat.Games {
		if isBIOS(game) || len(game.ROMs) == 0 {
			continue
		}
		nonBIOS++

		// Apply filter
		rom := game.ROMs[0]
		baseName := scraper.BaseName(rom.Name)
		ctx := scraper.BuildFilterContext(baseName, filterConfig)
		if shouldScrape, err := filter.ShouldScrape(ctx); err == nil && shouldScrape {
			toScrape++
		}
	}

	filteredOut := nonBIOS - toScrape

	if jsonOutput {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"name":         dat.Header.Name,
			"description":  dat.Header.Description,
			"version":      dat.Header.Version,
			"total_games":  stats.TotalGames,
			"total_roms":   stats.TotalROMs,
			"bios_count":   stats.BIOSCount,
			"verified":     stats.Verified,
			"with_serial":  stats.WithSerial,
			"with_header":  stats.WithHeader,
			"filter":       filter.Expression(),
			"filtered_out": filteredOut,
			"would_scrape": toScrape,
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("DAT File: %s\n", dat.Header.Name)
		fmt.Printf("Version: %s\n", dat.Header.Version)
		fmt.Printf("\n")
		fmt.Printf("Total Games: %d\n", stats.TotalGames)
		fmt.Printf("BIOS Entries: %d (will be skipped)\n", stats.BIOSCount)
		fmt.Printf("Non-BIOS Games: %d\n", nonBIOS)
		if filterExpr != "true" {
			fmt.Printf("\n")
			fmt.Printf("Filter: %s\n", filter.Expression())
			fmt.Printf("Filtered Out: %d\n", filteredOut)
		}
		fmt.Printf("Games to Scrape: %d\n", toScrape)
		fmt.Printf("\n")
		fmt.Printf("ROMs with Serial: %d\n", stats.WithSerial)
		fmt.Printf("ROMs with Header: %d\n", stats.WithHeader)
		fmt.Printf("Verified ROMs: %d\n", stats.Verified)
	}

	return nil
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func normalizeGamelistPath(path string) string {
	if path == "" {
		return ""
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return filepath.Join(path, "gamelist.xml")
	}
	if !strings.HasSuffix(path, ".xml") {
		return filepath.Join(path, "gamelist.xml")
	}
	return path
}

// isBIOS returns true if this is a BIOS entry (should be skipped)
func isBIOS(g datfile.Game) bool {
	return g.IsBIOS || strings.Contains(g.Name, "[BIOS]")
}

// stats holds statistics about a DAT file
type stats struct {
	TotalGames int
	TotalROMs  int
	BIOSCount  int
	Verified   int
	WithSerial int
	WithHeader int
}

// getStats calculates statistics for a DAT file
func getStats(f *datfile.Datafile) stats {
	s := stats{
		TotalGames: len(f.Games),
	}

	for _, g := range f.Games {
		if isBIOS(g) {
			s.BIOSCount++
		}
		for _, r := range g.ROMs {
			s.TotalROMs++
			if r.Status == "verified" {
				s.Verified++
			}
			if r.Serial != "" {
				s.WithSerial++
			}
			if r.Header != "" {
				s.WithHeader++
			}
		}
	}

	return s
}
