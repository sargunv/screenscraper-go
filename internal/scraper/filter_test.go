package scraper

import (
	"testing"
)

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{"true literal", "true", false},
		{"false literal", "false", false},
		{"simple metadata check", "missing.metadata", false},
		{"simple cover check", "missing.covers", false},
		{"or expression", "missing.metadata or missing.covers", false},
		{"and expression", "missing.metadata and missing.covers", false},
		{"not expression", "not missing.metadata", false},
		{"complex expression", "(missing.metadata or missing.covers) and not missing.videos", false},
		{"invalid field", "missing.invalid", true},
		{"syntax error", "missing.metadata +", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFilter(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilter_ShouldScrape(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		ctx        FilterContext
		want       bool
	}{
		{
			name:       "true always matches",
			expression: "true",
			ctx:        FilterContext{},
			want:       true,
		},
		{
			name:       "false never matches",
			expression: "false",
			ctx:        FilterContext{},
			want:       false,
		},
		{
			name:       "missing.metadata when metadata missing",
			expression: "missing.metadata",
			ctx:        FilterContext{Missing: MissingContext{Metadata: true}},
			want:       true,
		},
		{
			name:       "missing.metadata when metadata exists",
			expression: "missing.metadata",
			ctx:        FilterContext{Missing: MissingContext{Metadata: false}},
			want:       false,
		},
		{
			name:       "or expression - first true",
			expression: "missing.metadata or missing.covers",
			ctx:        FilterContext{Missing: MissingContext{Metadata: true, Covers: false}},
			want:       true,
		},
		{
			name:       "or expression - second true",
			expression: "missing.metadata or missing.covers",
			ctx:        FilterContext{Missing: MissingContext{Metadata: false, Covers: true}},
			want:       true,
		},
		{
			name:       "or expression - both false",
			expression: "missing.metadata or missing.covers",
			ctx:        FilterContext{Missing: MissingContext{Metadata: false, Covers: false}},
			want:       false,
		},
		{
			name:       "and expression - both true",
			expression: "missing.metadata and missing.covers",
			ctx:        FilterContext{Missing: MissingContext{Metadata: true, Covers: true}},
			want:       true,
		},
		{
			name:       "and expression - one false",
			expression: "missing.metadata and missing.covers",
			ctx:        FilterContext{Missing: MissingContext{Metadata: true, Covers: false}},
			want:       false,
		},
		{
			name:       "not expression",
			expression: "not missing.metadata",
			ctx:        FilterContext{Missing: MissingContext{Metadata: true}},
			want:       false,
		},
		{
			name:       "complex expression",
			expression: "not missing.metadata and missing.covers",
			ctx:        FilterContext{Missing: MissingContext{Metadata: false, Covers: true}},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFilter(tt.expression)
			if err != nil {
				t.Fatalf("NewFilter() error = %v", err)
			}

			got, err := f.ShouldScrape(tt.ctx)
			if err != nil {
				t.Fatalf("ShouldScrape() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("ShouldScrape() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFilterContext(t *testing.T) {
	// Test with nil config returns all missing
	ctx := BuildFilterContext("test-game", nil)
	if !ctx.Missing.Metadata {
		t.Error("expected Metadata to be missing with nil config")
	}
	if !ctx.Missing.Screenshots {
		t.Error("expected Screenshots to be missing with nil config")
	}

	// Test with empty config
	config := &FilterConfig{}
	ctx = BuildFilterContext("test-game", config)
	if !ctx.Missing.Metadata {
		t.Error("expected Metadata to be missing with empty config")
	}

	// Test with gamelist map (with extension)
	config = &FilterConfig{
		GamelistMap: map[string]bool{
			"./test-game.zip": true,
		},
	}
	ctx = BuildFilterContext("test-game", config)
	if ctx.Missing.Metadata {
		t.Error("expected Metadata to NOT be missing when in gamelist (with extension)")
	}

	// Test with gamelist map (ES-DE style, without extension)
	config = &FilterConfig{
		GamelistMap: map[string]bool{
			"./test-game": true,
		},
	}
	ctx = BuildFilterContext("test-game", config)
	if ctx.Missing.Metadata {
		t.Error("expected Metadata to NOT be missing when in gamelist (without extension)")
	}

	// Test game not in gamelist
	config = &FilterConfig{
		GamelistMap: map[string]bool{
			"./test-game": true,
		},
	}
	ctx = BuildFilterContext("other-game", config)
	if !ctx.Missing.Metadata {
		t.Error("expected Metadata to be missing when not in gamelist")
	}
}
