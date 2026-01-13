package gba

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

func TestIdentify(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "AGB_Rogue.gba")

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	ident, err := Identify(file, stat.Size())
	if err != nil {
		t.Fatalf("Identify() error = %v", err)
	}

	if ident.Platform != core.PlatformGBA {
		t.Errorf("Expected platform %s, got %s", core.PlatformGBA, ident.Platform)
	}

	if ident.TitleID != "AAAA" {
		t.Errorf("Expected title ID 'AAAA', got '%s'", ident.TitleID)
	}

	if ident.Title != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", ident.Title)
	}

	if ident.MakerCode != "AA" {
		t.Errorf("Expected maker code 'AA', got '%s'", ident.MakerCode)
	}
}
