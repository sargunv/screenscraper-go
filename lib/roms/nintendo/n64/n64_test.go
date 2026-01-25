package n64

import (
	"os"
	"testing"
)

func TestParseLittleEndian(t *testing.T) {
	romPath := "testdata/flames.n64"

	file, err := os.Open(romPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	info, err := Parse(file, stat.Size())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if info.ByteOrder != ByteOrderLittleEndian {
		t.Errorf("Expected byte order n64, got %s", info.ByteOrder)
	}
}
