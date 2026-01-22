package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sargunv/rom-tools/internal/cache"
	"github.com/sargunv/rom-tools/internal/region"
	"github.com/sargunv/rom-tools/lib/screenscraper"
)

// Worker handles scraping tasks
type Worker struct {
	id          int
	client      *screenscraper.ScreenscraperClient
	cache       *cache.DiskCache
	config      *Config
	rateLimiter *RateLimiter
	dedup       *Deduplicator
	updates     chan<- ProgressUpdate
}

// NewWorker creates a new worker
func NewWorker(id int, client *screenscraper.ScreenscraperClient, cache *cache.DiskCache, config *Config, rateLimiter *RateLimiter, dedup *Deduplicator, updates chan<- ProgressUpdate) *Worker {
	return &Worker{
		id:          id,
		client:      client,
		cache:       cache,
		config:      config,
		rateLimiter: rateLimiter,
		dedup:       dedup,
		updates:     updates,
	}
}

// Process handles a single lookup entry
func (w *Worker) Process(ctx context.Context, entry *LookupEntry) *ScrapeResult {
	result := &ScrapeResult{
		Entry: entry,
		Media: make(map[string]string),
	}

	mediaTotal := len(w.config.MediaTypes)

	// Notify progress
	w.sendUpdate(ProgressUpdate{
		Type:       UpdateTypeStarted,
		EntryName:  entry.Name,
		WorkerID:   w.id,
		MediaTotal: mediaTotal,
	})

	// Look up game info
	game, cached, notFound, err := w.lookupGame(ctx, entry)
	if err != nil {
		result.Error = err
		w.sendUpdate(ProgressUpdate{
			Type:       UpdateTypeError,
			EntryName:  entry.Name,
			WorkerID:   w.id,
			MediaTotal: mediaTotal,
			Error:      err,
		})
		return result
	}

	if notFound {
		result.Error = nil
		w.sendUpdate(ProgressUpdate{
			Type:       UpdateTypeNotFound,
			EntryName:  entry.Name,
			WorkerID:   w.id,
			MediaTotal: mediaTotal,
		})
		return result
	}

	result.Game = game
	result.Cached = cached

	// Download media
	mediaDone := 0
	cacheHits := 0 // count of API calls avoided due to cache (game info + media)
	mediaFailed := 0
	mediaMissing := 0

	// Count game info lookup as cached if it was
	if cached {
		cacheHits++
	}

	for _, esdeType := range w.config.MediaTypes {
		// Send progress update showing what we're working on
		w.sendUpdate(ProgressUpdate{
			Type:         UpdateTypeProgress,
			EntryName:    entry.Name,
			WorkerID:     w.id,
			MediaTotal:   mediaTotal,
			MediaDone:    mediaDone,
			CacheHits:    cacheHits,
			MediaFailed:  mediaFailed,
			MediaMissing: mediaMissing,
			CurrentMedia: esdeType,
		})

		ssTypes, ok := MediaTypeMapping[esdeType]
		if !ok {
			mediaMissing++
			continue
		}

		// Try each screenscraper type in order (fallback)
		gotMedia := false
		wasCached := false
		hadError := false
		for _, ssType := range ssTypes {
			path, fromCache, err := w.downloadMedia(ctx, entry, game, esdeType, ssType)
			if err != nil {
				hadError = true
				continue
			}
			if path != "" {
				result.Media[esdeType] = path
				mediaDone++
				wasCached = fromCache
				gotMedia = true
				break // Got this media type, move to next
			}
		}

		if gotMedia && wasCached {
			cacheHits++
		}
		if !gotMedia {
			if hadError {
				mediaFailed++
			} else {
				mediaMissing++
			}
		}
	}

	w.sendUpdate(ProgressUpdate{
		Type:         UpdateTypeFound,
		EntryName:    entry.Name,
		WorkerID:     w.id,
		MediaTotal:   mediaTotal,
		MediaDone:    mediaDone,
		CacheHits:    cacheHits,
		MediaFailed:  mediaFailed,
		MediaMissing: mediaMissing,
	})

	result.CacheHits = cacheHits
	return result
}

// lookupGame fetches game info from cache or API
// Returns (game, cached, notFound, error)
func (w *Worker) lookupGame(ctx context.Context, entry *LookupEntry) (*screenscraper.Game, bool, bool, error) {
	cacheKey := entry.Hashes.CacheKey()

	// Check cache first
	if !w.config.SkipCacheRead {
		if data, ok := w.cache.GetGameInfo(w.config.SystemID, cacheKey); ok {
			var game screenscraper.Game
			if err := json.Unmarshal(data, &game); err == nil {
				return &game, true, false, nil
			}
		}
	}

	// Use deduplicator to coalesce identical requests
	dedupKey := fmt.Sprintf("game:%s:%s", w.config.SystemID, cacheKey)

	type lookupResult struct {
		game     *screenscraper.Game
		notFound bool
	}

	result, err := DoTyped(w.dedup, dedupKey, func() (*lookupResult, error) {
		game, notFound, err := w.fetchGameFromAPI(ctx, entry)
		if err != nil {
			return nil, err
		}
		return &lookupResult{game: game, notFound: notFound}, nil
	})

	if err != nil {
		return nil, false, false, err
	}

	if result.notFound {
		return nil, false, true, nil
	}

	// Cache the result
	if !w.config.SkipCacheWrite {
		if data, err := json.Marshal(result.game); err == nil {
			w.cache.SetGameInfo(w.config.SystemID, cacheKey, data)
		}
	}

	return result.game, false, false, nil
}

// fetchGameFromAPI fetches game info from Screenscraper API
// Returns (game, notFound, error)
func (w *Worker) fetchGameFromAPI(ctx context.Context, entry *LookupEntry) (*screenscraper.Game, bool, error) {
	// Acquire rate limiter
	if err := w.rateLimiter.Acquire(ctx); err != nil {
		return nil, false, err
	}
	defer w.rateLimiter.Release()

	// Prepare parameters
	romSize := strconv.FormatInt(entry.Size, 10)
	params := &screenscraper.GetGameInfoParams{
		SystemID: w.config.SystemID,
		Crc:      entry.Hashes.CRC32,
		Md5:      entry.Hashes.MD5,
		Sha1:     entry.Hashes.SHA1,
		ROMSize:  romSize,
		ROMName:  entry.FileName,
		ROMType:  "rom",
	}

	if entry.Serial != "" {
		params.SerialNumber = entry.Serial
	}

	resp, err := w.client.GetGameInfoWithResponse(ctx, params)
	if err != nil {
		return nil, false, err
	}

	// Check for rate limiting
	if screenscraper.IsRateLimited(resp) {
		w.rateLimiter.TriggerBackoff()
		return nil, false, fmt.Errorf("rate limited")
	}

	// Check for not found
	if screenscraper.IsNotFound(resp) {
		return nil, true, nil
	}

	// Check for other errors
	if !screenscraper.IsSuccess(resp) {
		return nil, false, fmt.Errorf("API error: HTTP %d", resp.StatusCode())
	}

	w.rateLimiter.ResetBackoff()

	// Extract game from response
	// Since Response and Game are now value types (not pointers), check for empty game ID
	if resp.JSON200 == nil || resp.JSON200.Response.Game.Id == "" {
		return nil, true, nil // No game data
	}

	return &resp.JSON200.Response.Game, false, nil
}

// downloadMedia downloads a specific media type for a game
// Returns (path, cached, error) where cached indicates if the media was served from cache
func (w *Worker) downloadMedia(ctx context.Context, entry *LookupEntry, game *screenscraper.Game, esdeType, ssType string) (string, bool, error) {
	// Find the best media match based on region
	candidates := make([]region.Media, 0)
	for _, m := range game.Media {
		if m.Type == ssType {
			candidates = append(candidates, region.Media{
				Type:   m.Type,
				Region: m.Region,
				URL:    m.Url,
				Format: m.Format,
			})
		}
	}

	if len(candidates) == 0 {
		return "", false, nil // No media available
	}

	media := region.SelectMedia(candidates, ssType, entry.Regions, w.config.PreferredRegions)
	if media == nil {
		return "", false, nil
	}

	// Determine extension
	ext := MediaExtensions[ssType]
	if media.Format != "" {
		ext = strings.ToLower(media.Format)
	}

	// Build output path
	relativePath := filepath.Join(esdeType, entry.BaseName+"."+ext)
	outputPath := filepath.Join(w.config.MediaOutputDir, relativePath)

	// Check if output file already exists (skip unless overwrite)
	if !w.config.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return relativePath, true, nil // Already exists on disk
		}
	}

	// Get game ID
	if game.Id == "" {
		return "", false, nil // No game ID
	}

	// Check cache for media data
	mediaRegion := media.Region
	var data []byte
	cached := false
	if !w.config.SkipCacheRead {
		if cachedData, cachedExt, ok := w.cache.GetMedia(w.config.SystemID, game.Id, ssType, mediaRegion); ok {
			// Check if this is a cached "no media available" marker
			if cachedExt == "nomedia" {
				return "", false, nil
			}
			data = cachedData
			ext = cachedExt
			relativePath = filepath.Join(esdeType, entry.BaseName+"."+ext)
			outputPath = filepath.Join(w.config.MediaOutputDir, relativePath)
			cached = true
		}
	}

	// Download if not in cache
	if data == nil {
		// Use deduplicator for downloads
		dedupKey := fmt.Sprintf("media:%s:%s:%s:%s", w.config.SystemID, game.Id, ssType, mediaRegion)
		downloaded, err := DoTyped(w.dedup, dedupKey, func() ([]byte, error) {
			return w.downloadMediaFromAPI(ctx, game.Id, ssType, mediaRegion, ext)
		})

		if err != nil {
			return "", false, err
		}
		data = downloaded
	}

	if data == nil || len(data) == 0 {
		return "", false, nil // No data
	}

	// Write to output directory
	if w.config.MediaOutputDir != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return "", false, fmt.Errorf("failed to create media directory: %w", err)
		}
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return "", false, fmt.Errorf("failed to write media file: %w", err)
		}
	}

	return relativePath, cached, nil
}

// downloadMediaFromAPI downloads media from Screenscraper API
func (w *Worker) downloadMediaFromAPI(ctx context.Context, gameID, mediaType, mediaRegion, ext string) ([]byte, error) {
	// Acquire rate limiter
	if err := w.rateLimiter.Acquire(ctx); err != nil {
		return nil, err
	}
	defer w.rateLimiter.Release()

	// Build media identifier (e.g., "box-2D(us)")
	mediaID := mediaType
	if mediaRegion != "" {
		mediaID = fmt.Sprintf("%s(%s)", mediaType, mediaRegion)
	}

	params := &screenscraper.DownloadGameMediaParams{
		SystemID: w.config.SystemID,
		GameID:   gameID,
		Media:    mediaID,
	}

	resp, err := w.client.DownloadGameMediaWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}

	// Check for rate limiting
	if screenscraper.IsRateLimited(resp) {
		w.rateLimiter.TriggerBackoff()
		return nil, fmt.Errorf("rate limited")
	}

	// Check for errors (not found is not an error for media - just means no media)
	if !screenscraper.IsSuccess(resp) {
		// Media doesn't exist - cache this so we don't retry
		if !w.config.SkipCacheWrite {
			w.cache.SetMedia(w.config.SystemID, gameID, mediaType, mediaRegion, []byte("NOMEDIA"), "nomedia")
		}
		return nil, nil
	}

	data := resp.Body

	// Check for special responses
	dataStr := string(data)
	if dataStr == "NOMEDIA" || dataStr == "CRCOK" || dataStr == "MD5OK" || dataStr == "SHA1OK" {
		// Media doesn't exist or already up to date - cache this so we don't retry
		if !w.config.SkipCacheWrite {
			w.cache.SetMedia(w.config.SystemID, gameID, mediaType, mediaRegion, []byte("NOMEDIA"), "nomedia")
		}
		return nil, nil
	}

	w.rateLimiter.ResetBackoff()

	// Cache the media
	if !w.config.SkipCacheWrite {
		w.cache.SetMedia(w.config.SystemID, gameID, mediaType, mediaRegion, data, ext)
	}

	return data, nil
}

func (w *Worker) sendUpdate(update ProgressUpdate) {
	w.updates <- update
}

// ProgressUpdate represents a progress update from a worker
type ProgressUpdate struct {
	Type         UpdateType
	EntryName    string
	WorkerID     int
	MediaTotal   int    // total media types to download
	MediaDone    int    // media types downloaded successfully
	CacheHits    int    // API calls avoided due to cache (game info + media)
	MediaFailed  int    // media types that failed (error/timeout)
	MediaMissing int    // media types not available
	CurrentMedia string // currently downloading (for display)
	Error        error
}

// UpdateType represents the type of progress update
type UpdateType int

const (
	UpdateTypeStarted  UpdateType = iota
	UpdateTypeProgress            // media download progress
	UpdateTypeFound
	UpdateTypeNotFound
	UpdateTypeSkipped
	UpdateTypeError
)
