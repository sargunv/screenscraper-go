package xiso

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
	"github.com/sargunv/rom-tools/lib/romident/core"
)

func TestIdentify(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "xromwell.xiso.iso")

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

	if ident.Platform != core.PlatformXbox {
		t.Errorf("Expected platform %s, got %s", core.PlatformXbox, ident.Platform)
	}

	if ident.Title != "Xromwell" {
		t.Errorf("Expected title 'Xromwell', got '%s'", ident.Title)
	}
}
