package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sargunv/rom-tools/internal/cache"
	"github.com/sargunv/rom-tools/internal/region"
	"github.com/sargunv/rom-tools/lib/datfile"
	"github.com/sargunv/rom-tools/lib/screenscraper"
)

// Scraper orchestrates the scraping process
type Scraper struct {
	client      *screenscraper.ScreenscraperClient
	cache       *cache.DiskCache
	config      *Config
	rateLimiter *RateLimiter
	dedup       *Deduplicator

	// Progress tracking
	updates chan ProgressUpdate
}

// New creates a new scraper
func New(client *screenscraper.ScreenscraperClient, diskCache *cache.DiskCache, config *Config) *Scraper {
	return &Scraper{
		client:      client,
		cache:       diskCache,
		config:      config,
		rateLimiter: NewRateLimiter(config.MaxThreads, config.MaxRequestsPerMin),
		dedup:       NewDeduplicator(),
		updates:     make(chan ProgressUpdate, 100),
	}
}

// Updates returns the progress update channel
func (s *Scraper) Updates() <-chan ProgressUpdate {
	return s.updates
}

// RateLimiterStats returns current rate limiter statistics
func (s *Scraper) RateLimiterStats() RateLimiterStats {
	return s.rateLimiter.Stats()
}

// ScrapeResults contains the results of a scrape operation
type ScrapeResults struct {
	Results         []*ScrapeResult
	TotalEntries    int
	Found           int
	NotFound        int
	Skipped         int
	Errors          int
	MediaTotal      int
	MediaDownloaded int
	CacheHits       int
	FilteredOut     int // entries excluded by --filter expression
}

// ScrapeFromDAT scrapes games from a DAT file
func (s *Scraper) ScrapeFromDAT(ctx context.Context, datPath string) (*ScrapeResults, error) {
	// Parse DAT file
	dat, err := datfile.Parse(datPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DAT file: %w", err)
	}

	// Convert DAT entries to lookup entries (applies filter)
	entries, filteredOut := s.datToLookupEntries(dat)

	// Run scraping
	results, err := s.scrape(ctx, entries)
	if results != nil {
		results.FilteredOut = filteredOut
	}
	return results, err
}

// datToLookupEntries converts DAT games to lookup entries
// Returns entries to scrape and count of entries filtered out
func (s *Scraper) datToLookupEntries(dat *datfile.Datafile) ([]*LookupEntry, int) {
	var entries []*LookupEntry
	filteredOut := 0

	for _, game := range dat.Games {
		// Skip BIOS entries
		if isBIOS(game) {
			continue
		}

		// Use the first ROM (most DAT entries have one ROM)
		if len(game.ROMs) == 0 {
			continue
		}
		rom := game.ROMs[0]

		// Parse regions from filename
		regions := region.ParseFilename(game.Name)

		// Extract base name (without extension)
		baseName := BaseName(rom.Name)

		// Apply filter if configured
		if s.config.Filter != nil && s.config.FilterConfig != nil {
			ctx := BuildFilterContext(baseName, s.config.FilterConfig)
			shouldScrape, err := s.config.Filter.ShouldScrape(ctx)
			if err != nil {
				// On error, include the entry (fail open)
				shouldScrape = true
			}
			if !shouldScrape {
				filteredOut++
				continue
			}
		}

		entry := &LookupEntry{
			Name:     game.Name,
			FileName: rom.Name,
			Hashes: Hashes{
				SHA1:  rom.SHA1,
				MD5:   rom.MD5,
				CRC32: rom.CRC,
			},
			Serial:   rom.Serial,
			Size:     rom.Size,
			Regions:  regions,
			BaseName: baseName,
			Source:   SourceDAT,
		}

		entries = append(entries, entry)
	}

	return entries, filteredOut
}

// scrape runs the scraping operation on a list of entries
func (s *Scraper) scrape(ctx context.Context, entries []*LookupEntry) (*ScrapeResults, error) {
	results := &ScrapeResults{
		TotalEntries: len(entries),
		Results:      make([]*ScrapeResult, 0, len(entries)),
	}

	if len(entries) == 0 {
		close(s.updates)
		return results, nil
	}

	// Create entry channel
	entryChan := make(chan *LookupEntry, len(entries))
	for _, entry := range entries {
		entryChan <- entry
	}
	close(entryChan)

	// Create result channel
	resultChan := make(chan *ScrapeResult, len(entries))

	// Start workers
	var wg sync.WaitGroup
	numWorkers := s.config.MaxThreads
	if numWorkers > len(entries) {
		numWorkers = len(entries)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker := NewWorker(workerID, s.client, s.cache, s.config, s.rateLimiter, s.dedup, s.updates)

			for entry := range entryChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := worker.Process(ctx, entry)
					resultChan <- result
				}
			}
		}(i)
	}

	// Wait for workers and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
		close(s.updates)
	}()

	// Collect results
	for result := range resultChan {
		results.Results = append(results.Results, result)

		if result.Skipped {
			results.Skipped++
		} else if result.Error != nil {
			results.Errors++
		} else if result.Game != nil {
			results.Found++
			results.MediaDownloaded += len(result.Media)
		} else {
			results.NotFound++
		}

		results.MediaTotal += len(s.config.MediaTypes)
		results.CacheHits += result.CacheHits
	}

	return results, nil
}

// GetConfig returns the scraper configuration
func (s *Scraper) GetConfig() *Config {
	return s.config
}

// isBIOS returns true if this is a BIOS entry (should be skipped)
func isBIOS(g datfile.Game) bool {
	return g.IsBIOS || strings.Contains(g.Name, "[BIOS]")
}
