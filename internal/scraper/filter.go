package scraper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/sargunv/rom-tools/lib/esde"
)

// FilterContext contains the variables available in filter expressions
type FilterContext struct {
	Missing MissingContext `expr:"missing"`
}

// MissingContext tracks what's missing for a given entry
type MissingContext struct {
	Metadata      bool `expr:"metadata"`
	Screenshots   bool `expr:"screenshots"`
	Titlescreens  bool `expr:"titlescreens"`
	Covers        bool `expr:"covers"`
	Boxes3D       bool `expr:"3dboxes"`
	Marquees      bool `expr:"marquees"`
	Fanart        bool `expr:"fanart"`
	Videos        bool `expr:"videos"`
	Physicalmedia bool `expr:"physicalmedia"`
	Backcovers    bool `expr:"backcovers"`
}

// Filter evaluates filter expressions against entries
type Filter struct {
	program    *vm.Program
	expression string
}

// NewFilter creates a new filter from an expression string
// Example expressions:
//   - "true" (scrape all)
//   - "missing.metadata" (only scrape if not in gamelist)
//   - "missing.covers or missing.videos" (scrape if missing cover OR video)
func NewFilter(expression string) (*Filter, error) {
	program, err := expr.Compile(
		expression,
		expr.Env(FilterContext{}),
		expr.AsBool(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid filter expression: %w", err)
	}
	return &Filter{
		program:    program,
		expression: expression,
	}, nil
}

// Expression returns the original expression string
func (f *Filter) Expression() string {
	return f.expression
}

// ShouldScrape evaluates the filter for a given context
// Returns true if the entry should be scraped
func (f *Filter) ShouldScrape(ctx FilterContext) (bool, error) {
	result, err := expr.Run(f.program, ctx)
	if err != nil {
		return false, fmt.Errorf("filter evaluation failed: %w", err)
	}
	return result.(bool), nil
}

// FilterConfig holds configuration for filtering entries
type FilterConfig struct {
	GamelistPath string          // Path to existing gamelist.xml (for metadata checks)
	MediaDir     string          // Path to media directory (for media file checks)
	GamelistMap  map[string]bool // Map of ROM paths that exist in gamelist
}

// NewFilterConfig creates a FilterConfig from paths, loading gamelist if present
func NewFilterConfig(gamelistPath, mediaDir string) *FilterConfig {
	config := &FilterConfig{
		GamelistPath: gamelistPath,
		MediaDir:     mediaDir,
	}

	if gamelistPath != "" {
		if data, err := os.ReadFile(gamelistPath); err == nil {
			if gl, err := esde.Parse(data); err == nil {
				config.GamelistMap = make(map[string]bool)
				for _, game := range gl.Games {
					config.GamelistMap[game.Path] = true
				}
			}
		}
	}
	return config
}

// BuildFilterContext creates a FilterContext for a given entry
func BuildFilterContext(baseName string, config *FilterConfig) FilterContext {
	ctx := FilterContext{
		Missing: MissingContext{
			// Default to true (missing) - will be set to false if found
			Metadata:      true,
			Screenshots:   true,
			Titlescreens:  true,
			Covers:        true,
			Boxes3D:       true,
			Marquees:      true,
			Fanart:        true,
			Videos:        true,
			Physicalmedia: true,
			Backcovers:    true,
		},
	}

	if config == nil {
		return ctx
	}

	// Check if in gamelist
	if config.GamelistMap != nil {
		// First check without extension (ES-DE style)
		romPath := "./" + baseName
		if config.GamelistMap[romPath] {
			ctx.Missing.Metadata = false
		} else {
			// Also check common ROM extensions for compatibility
			for _, ext := range []string{".zip", ".7z", ".nes", ".sfc", ".smc", ".gba", ".gb", ".gbc", ".n64", ".z64", ".nds", ".md", ".gen", ".sms", ".gg"} {
				romPath := "./" + baseName + ext
				if config.GamelistMap[romPath] {
					ctx.Missing.Metadata = false
					break
				}
			}
		}
	}

	// Check media files
	if config.MediaDir != "" {
		ctx.Missing.Screenshots = !mediaExists(config.MediaDir, baseName, "screenshots")
		ctx.Missing.Titlescreens = !mediaExists(config.MediaDir, baseName, "titlescreens")
		ctx.Missing.Covers = !mediaExists(config.MediaDir, baseName, "covers")
		ctx.Missing.Boxes3D = !mediaExists(config.MediaDir, baseName, "3dboxes")
		ctx.Missing.Marquees = !mediaExists(config.MediaDir, baseName, "marquees")
		ctx.Missing.Fanart = !mediaExists(config.MediaDir, baseName, "fanart")
		ctx.Missing.Videos = !mediaExists(config.MediaDir, baseName, "videos")
		ctx.Missing.Physicalmedia = !mediaExists(config.MediaDir, baseName, "physicalmedia")
		ctx.Missing.Backcovers = !mediaExists(config.MediaDir, baseName, "backcovers")
	}

	return ctx
}

// mediaExtensions maps media type folders to possible file extensions
var mediaExtensions = map[string][]string{
	"screenshots":   {"png", "jpg", "jpeg"},
	"titlescreens":  {"png", "jpg", "jpeg"},
	"covers":        {"png", "jpg", "jpeg"},
	"3dboxes":       {"png", "jpg", "jpeg"},
	"marquees":      {"png", "jpg", "jpeg"},
	"fanart":        {"png", "jpg", "jpeg"},
	"videos":        {"mp4", "mkv", "avi", "webm"},
	"physicalmedia": {"png", "jpg", "jpeg"},
	"backcovers":    {"png", "jpg", "jpeg"},
}

// mediaExists checks if a media file exists for the given entry
func mediaExists(mediaDir, baseName, mediaType string) bool {
	extensions, ok := mediaExtensions[mediaType]
	if !ok {
		return false
	}

	for _, ext := range extensions {
		path := filepath.Join(mediaDir, mediaType, baseName+"."+ext)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
