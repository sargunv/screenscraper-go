package format

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestDetectByMagic(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		filename string
		want     Format
	}{
		{
			name:     "CHD file",
			filename: "empty.chd",
			want:     CHD,
		},
		{
			name:     "ZIP file",
			filename: "gbtictac.gb.zip",
			want:     ZIP,
		},
		{
			name:     "XISO file",
			filename: "xromwell.xiso.iso",
			want:     XISO,
		},
		{
			name:     "XBE file",
			filename: "xromwell/default.xbe",
			want:     XBE,
		},
		{
			name:     "GBA file",
			filename: "AGB_Rogue.gba",
			want:     GBA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(testutil.ROMsPath(t), tt.filename)

			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}

			got, err := detector.DetectByMagic(file, stat.Size())
			if err != nil {
				t.Fatalf("DetectByMagic() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("DetectByMagic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectByExtension(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		filename string
		want     Format
	}{
		{
			name:     "CHD extension",
			filename: "test.chd",
			want:     CHD,
		},
		{
			name:     "ZIP extension",
			filename: "test.zip",
			want:     ZIP,
		},
		{
			name:     "XISO extension",
			filename: "test.xiso",
			want:     XISO,
		},
		{
			name:     "XBE extension",
			filename: "default.xbe",
			want:     XBE,
		},
		{
			name:     "GBA extension",
			filename: "test.gba",
			want:     GBA,
		},
		{
			name:     "ISO extension (ambiguous)",
			filename: "test.iso",
			want:     Unknown,
		},
		{
			name:     "Unknown extension",
			filename: "test.unknown",
			want:     Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectByExtension(tt.filename)
			if got != tt.want {
				t.Errorf("DetectByExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
