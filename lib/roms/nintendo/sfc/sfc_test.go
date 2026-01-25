package sfc

import (
	"bytes"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	romPath := "testdata/col15.sfc"

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

	// Verify the file was parsed without error
	if info.Title == "" {
		t.Errorf("Expected non-empty title")
	}
}

func TestParse_Fields(t *testing.T) {
	romPath := "testdata/col15.sfc"

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

	// col15.sfc has title "32,768 color demo" and copier header
	if info.Title != "32,768 color demo" {
		t.Errorf("Title = %q, want %q", info.Title, "32,768 color demo")
	}
	if !info.HasCopierHeader {
		t.Errorf("HasCopierHeader = false, want true")
	}
	// Verify checksum validation passed
	if info.Checksum+info.ComplementCheck != 0xFFFF {
		t.Errorf("Checksum validation failed: 0x%04X + 0x%04X != 0xFFFF",
			info.Checksum, info.ComplementCheck)
	}
}

// makeSyntheticSNES creates a synthetic SNES ROM with a valid LoROM header.
func makeSyntheticSNES(title string, mapMode MapMode, destination Destination, cartType CartridgeType) []byte {
	// Create a LoROM-sized ROM (32KB minimum)
	rom := make([]byte, snesLoROMOffset+snesHeaderSize)

	// Write header at LoROM offset (0x7FC0)
	header := rom[snesLoROMOffset:]

	// Title (21 bytes, space-padded)
	titleBytes := []byte(title)
	if len(titleBytes) > snesTitleLen {
		titleBytes = titleBytes[:snesTitleLen]
	}
	copy(header[snesTitleOffset:], titleBytes)
	for i := len(titleBytes); i < snesTitleLen; i++ {
		header[snesTitleOffset+i] = ' '
	}

	// Map mode
	header[snesMapModeOffset] = byte(mapMode)

	// Cartridge type
	header[snesCartTypeOffset] = byte(cartType)

	// ROM size: 5 = 32KB (must be <= file size * 2 for validation)
	header[snesROMSizeOffset] = 5

	// RAM size: 0 = no RAM
	header[snesRAMSizeOffset] = 0

	// Destination
	header[snesDestCodeOffset] = byte(destination)

	// Old maker code
	header[snesMakerOldOffset] = 0x00

	// Version
	header[snesVersionOffset] = 0

	// Checksum: 0x0000, complement: 0xFFFF (valid pair)
	header[snesChecksumCOffset] = 0xFF
	header[snesChecksumCOffset+1] = 0xFF
	header[snesChecksumOffset] = 0x00
	header[snesChecksumOffset+1] = 0x00

	return rom
}

func TestParse_Synthetic(t *testing.T) {
	tests := []struct {
		name            string
		title           string
		mapMode         MapMode
		destination     Destination
		cartType        CartridgeType
		wantTitle       string
		wantMapMode     MapMode
		wantDestination Destination
		wantCartType    CartridgeType
	}{
		{
			name:            "LoROM Japan",
			title:           "TEST GAME",
			mapMode:         MapModeLoROM,
			destination:     DestinationJapan,
			cartType:        CartridgeROMOnly,
			wantTitle:       "TEST GAME",
			wantMapMode:     MapModeLoROM,
			wantDestination: DestinationJapan,
			wantCartType:    CartridgeROMOnly,
		},
		{
			name:            "HiROM USA with battery",
			title:           "BATTERY SAVE",
			mapMode:         MapModeHiROM,
			destination:     DestinationUSA,
			cartType:        CartridgeROMRAMBattery,
			wantTitle:       "BATTERY SAVE",
			wantMapMode:     MapModeHiROM,
			wantDestination: DestinationUSA,
			wantCartType:    CartridgeROMRAMBattery,
		},
		{
			name:            "FastROM Europe",
			title:           "FAST GAME EU",
			mapMode:         MapModeFastROMLoROM,
			destination:     DestinationEurope,
			cartType:        CartridgeROMRAM,
			wantTitle:       "FAST GAME EU",
			wantMapMode:     MapModeFastROMLoROM,
			wantDestination: DestinationEurope,
			wantCartType:    CartridgeROMRAM,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rom := makeSyntheticSNES(tc.title, tc.mapMode, tc.destination, tc.cartType)
			reader := bytes.NewReader(rom)

			info, err := Parse(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if info.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", info.Title, tc.wantTitle)
			}
			if info.MapMode != tc.wantMapMode {
				t.Errorf("MapMode = 0x%02X, want 0x%02X", info.MapMode, tc.wantMapMode)
			}
			if info.Destination != tc.wantDestination {
				t.Errorf("Destination = 0x%02X, want 0x%02X", info.Destination, tc.wantDestination)
			}
			if info.CartridgeType != tc.wantCartType {
				t.Errorf("CartridgeType = 0x%02X, want 0x%02X", info.CartridgeType, tc.wantCartType)
			}
			if info.HasCopierHeader {
				t.Errorf("HasCopierHeader = true, want false")
			}
		})
	}
}

func TestParse_InvalidHeader(t *testing.T) {
	// Create a ROM with invalid checksum
	rom := make([]byte, snesLoROMOffset+snesHeaderSize)
	header := rom[snesLoROMOffset:]

	// Set some title
	copy(header[snesTitleOffset:], []byte("INVALID GAME"))

	// Valid map mode
	header[snesMapModeOffset] = byte(MapModeLoROM)

	// Invalid checksum pair (doesn't sum to 0xFFFF)
	header[snesChecksumCOffset] = 0x00
	header[snesChecksumCOffset+1] = 0x00
	header[snesChecksumOffset] = 0x00
	header[snesChecksumOffset+1] = 0x00

	reader := bytes.NewReader(rom)
	_, err := Parse(reader, int64(len(rom)))
	if err == nil {
		t.Error("Parse() expected error for invalid checksum, got nil")
	}
}

func TestParse_TooSmall(t *testing.T) {
	// File too small for any header location
	data := make([]byte, 100)
	reader := bytes.NewReader(data)

	_, err := Parse(reader, int64(len(data)))
	if err == nil {
		t.Error("Parse() expected error for too small file, got nil")
	}
}
