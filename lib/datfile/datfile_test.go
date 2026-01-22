package datfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_NoIntro(t *testing.T) {
	path := filepath.Join("testdata", "Nintendo - Pokemon Mini (20250407-153358).dat")
	dat, err := Parse(path)
	if err != nil {
		t.Fatalf("failed to parse No-Intro DAT: %v", err)
	}

	// Verify header
	if dat.Header.Name != "Nintendo - Pokemon Mini" {
		t.Errorf("expected Name 'Nintendo - Pokemon Mini', got %q", dat.Header.Name)
	}
	if dat.Header.Version != "20250407-153358" {
		t.Errorf("expected Version '20250407-153358', got %q", dat.Header.Version)
	}
	if dat.Header.Homepage != "No-Intro" {
		t.Errorf("expected Homepage 'No-Intro', got %q", dat.Header.Homepage)
	}
	if dat.Header.URL != "https://www.no-intro.org" {
		t.Errorf("expected URL 'https://www.no-intro.org', got %q", dat.Header.URL)
	}
	if dat.Header.ID == nil || *dat.Header.ID != 14 {
		t.Errorf("expected ID 14, got %v", dat.Header.ID)
	}
	if dat.Header.ClrMamePro == nil {
		t.Error("expected ClrMamePro to be set")
	} else if dat.Header.ClrMamePro.ForceNoDump != "required" {
		t.Errorf("expected ForceNoDump 'required', got %q", dat.Header.ClrMamePro.ForceNoDump)
	}

	// Verify games exist
	if len(dat.Games) == 0 {
		t.Fatal("expected games, got none")
	}

	// Find the BIOS entry
	var biosGame *Game
	for i, g := range dat.Games {
		if g.Name == "[BIOS] Nintendo Pokemon Mini (World)" {
			biosGame = &dat.Games[i]
			break
		}
	}
	if biosGame == nil {
		t.Fatal("expected to find BIOS entry")
	}
	if biosGame.ID != "0001" {
		t.Errorf("expected BIOS ID '0001', got %q", biosGame.ID)
	}
	if len(biosGame.ROMs) != 1 {
		t.Fatalf("expected 1 ROM in BIOS, got %d", len(biosGame.ROMs))
	}
	if biosGame.ROMs[0].Size != 4096 {
		t.Errorf("expected ROM size 4096, got %d", biosGame.ROMs[0].Size)
	}
	if biosGame.ROMs[0].SHA256 != "45a1c7f28b9ad585e67f047abe9c1c956724bfcab8c9011002af4274e7c50e8f" {
		t.Errorf("unexpected SHA256: %s", biosGame.ROMs[0].SHA256)
	}

	// Find a game with clone info and category
	var cloneGame *Game
	for i, g := range dat.Games {
		if g.Name == "Lunch Time (Europe) (Demo) (Pokemon Channel)" {
			cloneGame = &dat.Games[i]
			break
		}
	}
	if cloneGame == nil {
		t.Fatal("expected to find clone game")
	}
	if cloneGame.CloneOfID != "0038" {
		t.Errorf("expected CloneOfID '0038', got %q", cloneGame.CloneOfID)
	}
	if len(cloneGame.Categories) != 1 || cloneGame.Categories[0] != "Games" {
		t.Errorf("expected Categories ['Games'], got %v", cloneGame.Categories)
	}
	if len(cloneGame.ROMs) != 1 || cloneGame.ROMs[0].Serial != "MLTE" {
		t.Error("expected ROM with serial MLTE")
	}

	// Find a game with verified status
	var verifiedGame *Game
	for i, g := range dat.Games {
		if g.Name == "Pokemon Party Mini (Japan)" {
			verifiedGame = &dat.Games[i]
			break
		}
	}
	if verifiedGame != nil && len(verifiedGame.ROMs) > 0 {
		if verifiedGame.ROMs[0].Status != "verified" {
			t.Errorf("expected verified status, got %q", verifiedGame.ROMs[0].Status)
		}
	}
}

func TestParse_Redump(t *testing.T) {
	path := filepath.Join("testdata", "Atari - Jaguar CD Interactive Multimedia System - Datfile (27) (2025-11-06 20-09-06).dat")
	dat, err := Parse(path)
	if err != nil {
		t.Fatalf("failed to parse Redump DAT: %v", err)
	}

	// Verify header
	if dat.Header.Name != "Atari - Jaguar CD Interactive Multimedia System" {
		t.Errorf("unexpected Name: %q", dat.Header.Name)
	}
	if dat.Header.Date != "2025-11-06 20-09-06" {
		t.Errorf("expected Date '2025-11-06 20-09-06', got %q", dat.Header.Date)
	}
	if dat.Header.Author != "redump.org" {
		t.Errorf("expected Author 'redump.org', got %q", dat.Header.Author)
	}

	// Verify games exist
	if len(dat.Games) == 0 {
		t.Fatal("expected games, got none")
	}

	// Find a game with multiple ROMs (typical for CD-based games)
	var multiRomGame *Game
	for i, g := range dat.Games {
		if g.Name == "Vid Grid (USA)" {
			multiRomGame = &dat.Games[i]
			break
		}
	}
	if multiRomGame == nil {
		t.Fatal("expected to find 'Vid Grid (USA)' game")
	}
	if len(multiRomGame.ROMs) < 10 {
		t.Errorf("expected multiple ROMs for CD game, got %d", len(multiRomGame.ROMs))
	}
	if len(multiRomGame.Categories) != 1 || multiRomGame.Categories[0] != "Games" {
		t.Errorf("expected Categories ['Games'], got %v", multiRomGame.Categories)
	}

	// Verify ROM details
	var cueFile *ROM
	for i, r := range multiRomGame.ROMs {
		if r.Name == "Vid Grid (USA).cue" {
			cueFile = &multiRomGame.ROMs[i]
			break
		}
	}
	if cueFile == nil {
		t.Fatal("expected to find cue file")
	}
	if cueFile.Size != 2094 {
		t.Errorf("expected cue size 2094, got %d", cueFile.Size)
	}
}

func TestParse_TOSEC(t *testing.T) {
	path := filepath.Join("testdata", "Sony PocketStation - Games (TOSEC-v2022-06-08_CM).dat")
	dat, err := Parse(path)
	if err != nil {
		t.Fatalf("failed to parse TOSEC DAT: %v", err)
	}

	// Verify header
	if dat.Header.Name != "Sony PocketStation - Games" {
		t.Errorf("unexpected Name: %q", dat.Header.Name)
	}
	if dat.Header.Category != "TOSEC" {
		t.Errorf("expected Category 'TOSEC', got %q", dat.Header.Category)
	}
	if dat.Header.Email != "contact@tosecdev.org" {
		t.Errorf("expected Email 'contact@tosecdev.org', got %q", dat.Header.Email)
	}
	if dat.Header.Homepage != "TOSEC" {
		t.Errorf("expected Homepage 'TOSEC', got %q", dat.Header.Homepage)
	}

	// Verify games exist
	if len(dat.Games) == 0 {
		t.Fatal("expected games, got none")
	}

	// Find the game
	var game *Game
	for i, g := range dat.Games {
		if g.Name == "Pocket Worm (2015)(Mihai, Sebastian)" {
			game = &dat.Games[i]
			break
		}
	}
	if game == nil {
		t.Fatal("expected to find 'Pocket Worm' game")
	}
	if len(game.ROMs) != 1 {
		t.Fatalf("expected 1 ROM, got %d", len(game.ROMs))
	}
	if game.ROMs[0].Size != 8192 {
		t.Errorf("expected size 8192, got %d", game.ROMs[0].Size)
	}
}

func TestParse_RetroAchievements(t *testing.T) {
	path := filepath.Join("testdata", "RA - NEC PC-FX.dat")
	dat, err := Parse(path)
	if err != nil {
		t.Fatalf("failed to parse RetroAchievements DAT: %v", err)
	}

	// Verify header
	if dat.Header.Name != "RA - NEC PC-FX" {
		t.Errorf("unexpected Name: %q", dat.Header.Name)
	}
	if dat.Header.Homepage != "https://retroachievements.org" {
		t.Errorf("unexpected Homepage: %q", dat.Header.Homepage)
	}
	if dat.Header.ClrMamePro == nil {
		t.Error("expected ClrMamePro to be set")
	} else if dat.Header.ClrMamePro.ForcePacking != "unzip" {
		t.Errorf("expected ForcePacking 'unzip', got %q", dat.Header.ClrMamePro.ForcePacking)
	}

	// RetroAchievements uses <machine> instead of <game>
	if len(dat.Games) == 0 {
		t.Fatal("expected games/machines, got none")
	}

	// Find a machine with disk and release entries
	var machine *Game
	for i, g := range dat.Games {
		if g.Name == "Battle Heat (Japan)" {
			machine = &dat.Games[i]
			break
		}
	}
	if machine == nil {
		t.Fatal("expected to find 'Battle Heat (Japan)' machine")
	}
	if len(machine.Categories) != 1 || machine.Categories[0] != "Retail" {
		t.Errorf("expected Categories ['Retail'], got %v", machine.Categories)
	}
	if len(machine.Releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(machine.Releases))
	}
	if machine.Releases[0].Region != "Japan" {
		t.Errorf("expected release region 'Japan', got %q", machine.Releases[0].Region)
	}
	if len(machine.Disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(machine.Disks))
	}
	if machine.Disks[0].SHA1 != "1a73ab8064ddfaebd5e555f51e4c153da6027fd7" {
		t.Errorf("unexpected disk SHA1: %s", machine.Disks[0].SHA1)
	}

	// Find a clone entry
	var clone *Game
	for i, g := range dat.Games {
		if g.Name == "Choujin Heiki Zeroigar (Japan) (En) (v1.0.1) (SamIAm)" {
			clone = &dat.Games[i]
			break
		}
	}
	if clone == nil {
		t.Fatal("expected to find clone entry")
	}
	if clone.CloneOf != "Choujin Heiki Zeroigar (Japan)" {
		t.Errorf("unexpected CloneOf: %q", clone.CloneOf)
	}
}

func TestParse_NoIntroTranslation(t *testing.T) {
	path := filepath.Join("testdata", "Nintendo - Wii [T-En] Collection (2025-10-01).dat")
	dat, err := Parse(path)
	if err != nil {
		t.Fatalf("failed to parse No-Intro Translation DAT: %v", err)
	}

	// Verify header
	if dat.Header.Name != "Nintendo - Wii [T-En] Collection" {
		t.Errorf("unexpected Name: %q", dat.Header.Name)
	}
	if dat.Header.Category != "Standard DatFile" {
		t.Errorf("expected Category 'Standard DatFile', got %q", dat.Header.Category)
	}
	if dat.Header.Comment != "-insert comment-" {
		t.Errorf("expected Comment '-insert comment-', got %q", dat.Header.Comment)
	}
	// Empty ClrMamePro element should still be parsed
	if dat.Header.ClrMamePro == nil {
		t.Error("expected ClrMamePro to be set (even if empty)")
	}

	// Uses <machine> elements
	if len(dat.Games) == 0 {
		t.Fatal("expected machines, got none")
	}

	// Find a machine with multiple ROMs
	var docsGame *Game
	for i, g := range dat.Games {
		if g.Name == "_Nintendo Wii [T-En] Docs" {
			docsGame = &dat.Games[i]
			break
		}
	}
	if docsGame == nil {
		t.Fatal("expected to find '_Nintendo Wii [T-En] Docs' machine")
	}
	if len(docsGame.ROMs) < 5 {
		t.Errorf("expected multiple ROMs in docs machine, got %d", len(docsGame.ROMs))
	}
}

func TestParse_MissingFile(t *testing.T) {
	_, err := Parse("nonexistent.dat")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParse_InvalidXML(t *testing.T) {
	// Create a temporary file with invalid XML
	tmpFile, err := os.CreateTemp("", "invalid-*.dat")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("this is not valid xml")
	tmpFile.Close()

	_, err = Parse(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestParse_EmptyDatafile(t *testing.T) {
	// Create a temporary file with valid but empty datafile
	tmpFile, err := os.CreateTemp("", "empty-*.dat")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(`<?xml version="1.0"?>
<datafile>
	<header>
		<name>Empty DAT</name>
	</header>
</datafile>`)
	tmpFile.Close()

	dat, err := Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to parse empty datafile: %v", err)
	}
	if dat.Header.Name != "Empty DAT" {
		t.Errorf("expected Name 'Empty DAT', got %q", dat.Header.Name)
	}
	if len(dat.Games) != 0 {
		t.Errorf("expected 0 games, got %d", len(dat.Games))
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"no", false},
		{"false", false},
		{"0", false},
		{"", false},
		{"  yes  ", true},
	}

	for _, test := range tests {
		result := parseBool(test.input)
		if result != test.expected {
			t.Errorf("parseBool(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}
