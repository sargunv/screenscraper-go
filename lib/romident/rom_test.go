package romident

import (
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestIdentifyLooseFile(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb")

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeDefault})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeFile {
		t.Errorf("Expected type %s, got %s", ROMTypeFile, rom.Type)
	}

	if len(rom.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(rom.Files))
	}

	file, ok := rom.Files["gbtictac.gb"]
	if !ok {
		t.Fatal("Expected file 'gbtictac.gb' not found")
	}

	if file.Size != 32768 {
		t.Errorf("Expected size 32768, got %d", file.Size)
	}

	if len(file.Hashes) != 3 {
		t.Fatalf("Expected 3 hashes, got %d", len(file.Hashes))
	}

	// Verify SHA1 hash
	var sha1Hash *Hash
	for i := range file.Hashes {
		if file.Hashes[i].Algorithm == HashSHA1 {
			sha1Hash = &file.Hashes[i]
			break
		}
	}
	if sha1Hash == nil {
		t.Fatal("SHA1 hash not found")
	}
	if sha1Hash.Value != "48a59d5b31e374731ece4d9eb33679d38143495e" {
		t.Errorf("Expected SHA1 '48a59d5b31e374731ece4d9eb33679d38143495e', got '%s'", sha1Hash.Value)
	}
	if sha1Hash.Source != "calculated" {
		t.Errorf("Expected SHA1 source 'calculated', got '%s'", sha1Hash.Source)
	}

	// Verify CRC32 hash
	var crc32Hash *Hash
	for i := range file.Hashes {
		if file.Hashes[i].Algorithm == HashCRC32 {
			crc32Hash = &file.Hashes[i]
			break
		}
	}
	if crc32Hash == nil {
		t.Fatal("CRC32 hash not found")
	}
	if crc32Hash.Value != "775ae755" {
		t.Errorf("Expected CRC32 '775ae755', got '%s'", crc32Hash.Value)
	}
}

func TestIdentifyZIP(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb.zip")

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeDefault})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeZIP {
		t.Errorf("Expected type %s, got %s", ROMTypeZIP, rom.Type)
	}

	if len(rom.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(rom.Files))
	}

	file, ok := rom.Files["gbtictac.gb"]
	if !ok {
		t.Fatal("Expected file 'gbtictac.gb' not found")
	}

	if file.Size != 32768 {
		t.Errorf("Expected size 32768, got %d", file.Size)
	}

	// In default mode, ZIP should only have CRC32 from metadata
	if len(file.Hashes) != 1 {
		t.Fatalf("Expected 1 hash (CRC32 from metadata), got %d", len(file.Hashes))
	}

	crc32Hash := file.Hashes[0]
	if crc32Hash.Algorithm != HashCRC32 {
		t.Errorf("Expected CRC32 hash, got %s", crc32Hash.Algorithm)
	}
	if crc32Hash.Value != "775ae755" {
		t.Errorf("Expected CRC32 '775ae755', got '%s'", crc32Hash.Value)
	}
	if crc32Hash.Source != "zip-metadata" {
		t.Errorf("Expected source 'zip-metadata', got '%s'", crc32Hash.Source)
	}
}

func TestIdentifyZIPSlowMode(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "gbtictac.gb.zip")

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeSlow})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeZIP {
		t.Errorf("Expected type %s, got %s", ROMTypeZIP, rom.Type)
	}

	if len(rom.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(rom.Files))
	}

	file, ok := rom.Files["gbtictac.gb"]
	if !ok {
		t.Fatal("Expected file 'gbtictac.gb' not found")
	}

	// In slow mode, ZIP should have full hashes calculated
	if len(file.Hashes) != 3 {
		t.Fatalf("Expected 3 hashes (SHA1, MD5, CRC32), got %d", len(file.Hashes))
	}

	// Verify SHA1 hash
	var sha1Hash *Hash
	for i := range file.Hashes {
		if file.Hashes[i].Algorithm == HashSHA1 {
			sha1Hash = &file.Hashes[i]
			break
		}
	}
	if sha1Hash == nil {
		t.Fatal("SHA1 hash not found")
	}
	if sha1Hash.Value != "48a59d5b31e374731ece4d9eb33679d38143495e" {
		t.Errorf("Expected SHA1 '48a59d5b31e374731ece4d9eb33679d38143495e', got '%s'", sha1Hash.Value)
	}
	if sha1Hash.Source != "calculated" {
		t.Errorf("Expected SHA1 source 'calculated', got '%s'", sha1Hash.Source)
	}
}

func TestIdentifyFolder(t *testing.T) {
	romPath := filepath.Join(testutil.ROMsPath(t), "xromwell")

	rom, err := IdentifyROM(romPath, Options{HashMode: HashModeDefault})
	if err != nil {
		t.Fatalf("IdentifyROM() error = %v", err)
	}

	if rom.Type != ROMTypeFolder {
		t.Errorf("Expected type %s, got %s", ROMTypeFolder, rom.Type)
	}

	if len(rom.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(rom.Files))
	}

	file, ok := rom.Files["default.xbe"]
	if !ok {
		t.Fatal("Expected file 'default.xbe' not found")
	}

	if file.Size != 290768 {
		t.Errorf("Expected size 290768, got %d", file.Size)
	}

	// Verify SHA1 hash
	var sha1Hash *Hash
	for i := range file.Hashes {
		if file.Hashes[i].Algorithm == HashSHA1 {
			sha1Hash = &file.Hashes[i]
			break
		}
	}
	if sha1Hash == nil {
		t.Fatal("SHA1 hash not found")
	}
	if sha1Hash.Value != "8f6c5fea086979c22b00b8e0bb957b40bc8b32d8" {
		t.Errorf("Expected SHA1 '8f6c5fea086979c22b00b8e0bb957b40bc8b32d8', got '%s'", sha1Hash.Value)
	}
}
