package gba

import (
	"bytes"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	romPath := "testdata/AGB_Rogue.gba"

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

	if info.Title != "ROGUE" {
		t.Errorf("Expected title 'ROGUE', got '%s'", info.Title)
	}

	if info.GameCode != "AAAA" {
		t.Errorf("Expected game code 'AAAA', got '%s'", info.GameCode)
	}

	if info.GameType != GameTypeNormalOld {
		t.Errorf("Expected game type 'A' (0x%02X), got 0x%02X", GameTypeNormalOld, info.GameType)
	}

	if info.Destination != Destination('A') {
		t.Errorf("Expected destination 'A' (0x%02X), got 0x%02X", 'A', info.Destination)
	}

	if info.MakerCode != "AA" {
		t.Errorf("Expected maker code 'AA', got '%s'", info.MakerCode)
	}

	if info.MainUnitCode != 0x00 {
		t.Errorf("Expected main unit code 0x00, got 0x%02X", info.MainUnitCode)
	}

	if info.DeviceType != 0x80 {
		t.Errorf("Expected device type 0x80, got 0x%02X", info.DeviceType)
	}

	if info.Version != 1 {
		t.Errorf("Expected version 1, got %d", info.Version)
	}

	if info.HeaderChecksum != 0xC8 {
		t.Errorf("Expected header checksum 0xC8, got 0x%02X", info.HeaderChecksum)
	}
}

// makeSyntheticGBA creates a minimal valid GBA ROM header for testing.
func makeSyntheticGBA(title, gameCode, makerCode string, mainUnit, deviceType, version, checksum byte) []byte {
	header := make([]byte, gbaHeaderSize)

	// Copy title (up to 12 bytes)
	copy(header[gbaTitleOffset:gbaTitleOffset+gbaTitleLen], title)

	// Copy game code (4 bytes)
	copy(header[gbaGameCodeOffset:gbaGameCodeOffset+gbaGameCodeLen], gameCode)

	// Copy maker code (2 bytes)
	copy(header[gbaMakerOffset:gbaMakerOffset+gbaMakerLen], makerCode)

	// Set fixed value
	header[gbaFixedOffset] = gbaFixedValue

	// Set hardware fields
	header[gbaMainUnitOffset] = mainUnit
	header[gbaDeviceTypeOffset] = deviceType

	// Set version and checksum
	header[gbaVersionOffset] = version
	header[gbaChecksumOffset] = checksum

	return header
}

func TestParseSyntheticGameTypes(t *testing.T) {
	tests := []struct {
		name         string
		gameCode     string
		wantGameType GameType
		wantDest     Destination
	}{
		{
			name:         "Normal old (A)",
			gameCode:     "AMGE",
			wantGameType: GameTypeNormalOld,
			wantDest:     DestinationUSA,
		},
		{
			name:         "Normal new (B)",
			gameCode:     "BPEJ",
			wantGameType: GameTypeNormalNew,
			wantDest:     DestinationJapan,
		},
		{
			name:         "Famicom Mini (F)",
			gameCode:     "FMRP",
			wantGameType: GameTypeFamicom,
			wantDest:     DestinationEurope,
		},
		{
			name:         "Acceleration sensor (K)",
			gameCode:     "KYGE",
			wantGameType: GameTypeAcceleration,
			wantDest:     DestinationUSA,
		},
		{
			name:         "e-Reader (P)",
			gameCode:     "PSAE",
			wantGameType: GameTypeEReader,
			wantDest:     DestinationUSA,
		},
		{
			name:         "Rumble + gyro (R)",
			gameCode:     "RZWJ",
			wantGameType: GameTypeRumbleGyro,
			wantDest:     DestinationJapan,
		},
		{
			name:         "RTC + solar (U)",
			gameCode:     "U3IJ",
			wantGameType: GameTypeRTCSolar,
			wantDest:     DestinationJapan,
		},
		{
			name:         "Rumble only (V)",
			gameCode:     "V49J",
			wantGameType: GameTypeRumble,
			wantDest:     DestinationJapan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rom := makeSyntheticGBA("TEST", tt.gameCode, "01", 0x00, 0x00, 0, 0)
			reader := bytes.NewReader(rom)

			info, err := Parse(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if info.GameType != tt.wantGameType {
				t.Errorf("GameType = 0x%02X, want 0x%02X", info.GameType, tt.wantGameType)
			}
			if info.Destination != tt.wantDest {
				t.Errorf("Destination = 0x%02X, want 0x%02X", info.Destination, tt.wantDest)
			}
		})
	}
}

func TestParseSyntheticDestinations(t *testing.T) {
	tests := []struct {
		name     string
		destCode byte
		wantDest Destination
	}{
		{"Japan", 'J', DestinationJapan},
		{"USA", 'E', DestinationUSA},
		{"Europe", 'P', DestinationEurope},
		{"France", 'F', DestinationFrance},
		{"Spain", 'S', DestinationSpain},
		{"Germany", 'D', DestinationGermany},
		{"Italy", 'I', DestinationItaly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameCode := "AXX" + string(tt.destCode)
			rom := makeSyntheticGBA("TEST", gameCode, "01", 0x00, 0x00, 0, 0)
			reader := bytes.NewReader(rom)

			info, err := Parse(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if info.Destination != tt.wantDest {
				t.Errorf("Destination = 0x%02X, want 0x%02X", info.Destination, tt.wantDest)
			}
		})
	}
}

func TestParseSyntheticHardwareFields(t *testing.T) {
	tests := []struct {
		name           string
		mainUnit       byte
		deviceType     byte
		version        byte
		checksum       byte
		wantMainUnit   byte
		wantDeviceType byte
		wantVersion    int
		wantChecksum   byte
	}{
		{
			name:           "Standard GBA",
			mainUnit:       0x00,
			deviceType:     0x00,
			version:        0,
			checksum:       0x00,
			wantMainUnit:   0x00,
			wantDeviceType: 0x00,
			wantVersion:    0,
			wantChecksum:   0x00,
		},
		{
			name:           "Debug hardware (DACS)",
			mainUnit:       0x00,
			deviceType:     0x80,
			version:        1,
			checksum:       0xAB,
			wantMainUnit:   0x00,
			wantDeviceType: 0x80,
			wantVersion:    1,
			wantChecksum:   0xAB,
		},
		{
			name:           "High version number",
			mainUnit:       0x00,
			deviceType:     0x00,
			version:        255,
			checksum:       0xFF,
			wantMainUnit:   0x00,
			wantDeviceType: 0x00,
			wantVersion:    255,
			wantChecksum:   0xFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rom := makeSyntheticGBA("TEST", "AXXX", "01", tt.mainUnit, tt.deviceType, tt.version, tt.checksum)
			reader := bytes.NewReader(rom)

			info, err := Parse(reader, int64(len(rom)))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if info.MainUnitCode != tt.wantMainUnit {
				t.Errorf("MainUnitCode = 0x%02X, want 0x%02X", info.MainUnitCode, tt.wantMainUnit)
			}
			if info.DeviceType != tt.wantDeviceType {
				t.Errorf("DeviceType = 0x%02X, want 0x%02X", info.DeviceType, tt.wantDeviceType)
			}
			if info.Version != tt.wantVersion {
				t.Errorf("Version = %d, want %d", info.Version, tt.wantVersion)
			}
			if info.HeaderChecksum != tt.wantChecksum {
				t.Errorf("HeaderChecksum = 0x%02X, want 0x%02X", info.HeaderChecksum, tt.wantChecksum)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	t.Run("file too small", func(t *testing.T) {
		rom := make([]byte, gbaHeaderSize-1)
		reader := bytes.NewReader(rom)

		_, err := Parse(reader, int64(len(rom)))
		if err == nil {
			t.Error("Expected error for file too small, got nil")
		}
	})

	t.Run("invalid fixed byte", func(t *testing.T) {
		rom := make([]byte, gbaHeaderSize)
		rom[gbaFixedOffset] = 0x00 // Invalid, should be 0x96
		reader := bytes.NewReader(rom)

		_, err := Parse(reader, int64(len(rom)))
		if err == nil {
			t.Error("Expected error for invalid fixed byte, got nil")
		}
	})
}
