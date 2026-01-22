package esde

import (
	"testing"
	"time"
)

func TestParseGameFullEntry(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<gameList>
  <game>
    <path>./test.gba</path>
    <name>Test Game</name>
    <desc>A test game description</desc>
    <image>./media/images/test.png</image>
    <thumbnail>./media/thumbnails/test.png</thumbnail>
    <video>./media/videos/test.mp4</video>
    <rating>0.85</rating>
    <releasedate>19910623T000000</releasedate>
    <developer>Test Developer</developer>
    <publisher>Test Publisher</publisher>
    <genre>Action, Adventure</genre>
    <players>2</players>
    <playcount>5</playcount>
    <lastplayed>20240115T143022</lastplayed>
  </game>
</gameList>`

	list, err := Parse([]byte(xml))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(list.Games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(list.Games))
	}

	g := list.Games[0]
	if g.Path != "./test.gba" {
		t.Errorf("Path: got %q, want %q", g.Path, "./test.gba")
	}
	if g.Name != "Test Game" {
		t.Errorf("Name: got %q, want %q", g.Name, "Test Game")
	}
	if g.Desc != "A test game description" {
		t.Errorf("Desc: got %q, want %q", g.Desc, "A test game description")
	}
	if g.Image != "./media/images/test.png" {
		t.Errorf("Image: got %q, want %q", g.Image, "./media/images/test.png")
	}
	if g.Thumbnail != "./media/thumbnails/test.png" {
		t.Errorf("Thumbnail: got %q, want %q", g.Thumbnail, "./media/thumbnails/test.png")
	}
	if g.Video != "./media/videos/test.mp4" {
		t.Errorf("Video: got %q, want %q", g.Video, "./media/videos/test.mp4")
	}
	if g.Rating != 0.85 {
		t.Errorf("Rating: got %v, want %v", g.Rating, 0.85)
	}
	expectedReleaseDate := time.Date(1991, 6, 23, 0, 0, 0, 0, time.UTC)
	if !g.ReleaseDate.Equal(expectedReleaseDate) {
		t.Errorf("ReleaseDate: got %v, want %v", g.ReleaseDate.Time, expectedReleaseDate)
	}
	if g.Developer != "Test Developer" {
		t.Errorf("Developer: got %q, want %q", g.Developer, "Test Developer")
	}
	if g.Publisher != "Test Publisher" {
		t.Errorf("Publisher: got %q, want %q", g.Publisher, "Test Publisher")
	}
	if g.Genre != "Action, Adventure" {
		t.Errorf("Genre: got %q, want %q", g.Genre, "Action, Adventure")
	}
	if g.Players != 2 {
		t.Errorf("Players: got %d, want %d", g.Players, 2)
	}
	if g.PlayCount != 5 {
		t.Errorf("PlayCount: got %d, want %d", g.PlayCount, 5)
	}
	expectedLastPlayed := time.Date(2024, 1, 15, 14, 30, 22, 0, time.UTC)
	if !g.LastPlayed.Equal(expectedLastPlayed) {
		t.Errorf("LastPlayed: got %v, want %v", g.LastPlayed.Time, expectedLastPlayed)
	}
}

func TestParseFolderEntry(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<gameList>
  <folder>
    <path>./My Games</path>
    <name>My Games</name>
    <desc>Collection of my favorite games</desc>
    <image>./media/images/mygames.png</image>
    <thumbnail>./media/thumbnails/mygames.png</thumbnail>
  </folder>
</gameList>`

	list, err := Parse([]byte(xml))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(list.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(list.Folders))
	}

	f := list.Folders[0]
	if f.Path != "./My Games" {
		t.Errorf("Path: got %q, want %q", f.Path, "./My Games")
	}
	if f.Name != "My Games" {
		t.Errorf("Name: got %q, want %q", f.Name, "My Games")
	}
	if f.Desc != "Collection of my favorite games" {
		t.Errorf("Desc: got %q, want %q", f.Desc, "Collection of my favorite games")
	}
	if f.Image != "./media/images/mygames.png" {
		t.Errorf("Image: got %q, want %q", f.Image, "./media/images/mygames.png")
	}
	if f.Thumbnail != "./media/thumbnails/mygames.png" {
		t.Errorf("Thumbnail: got %q, want %q", f.Thumbnail, "./media/thumbnails/mygames.png")
	}
}

func TestParseMixedGamesAndFolders(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<gameList>
  <folder>
    <path>./RPGs</path>
    <name>RPGs</name>
  </folder>
  <game>
    <path>./game1.gba</path>
    <name>Game One</name>
  </game>
  <folder>
    <path>./Action</path>
    <name>Action</name>
  </folder>
  <game>
    <path>./game2.gba</path>
    <name>Game Two</name>
  </game>
</gameList>`

	list, err := Parse([]byte(xml))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(list.Games) != 2 {
		t.Errorf("expected 2 games, got %d", len(list.Games))
	}
	if len(list.Folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(list.Folders))
	}

	if list.Games[0].Name != "Game One" {
		t.Errorf("first game name: got %q, want %q", list.Games[0].Name, "Game One")
	}
	if list.Games[1].Name != "Game Two" {
		t.Errorf("second game name: got %q, want %q", list.Games[1].Name, "Game Two")
	}
	if list.Folders[0].Name != "RPGs" {
		t.Errorf("first folder name: got %q, want %q", list.Folders[0].Name, "RPGs")
	}
	if list.Folders[1].Name != "Action" {
		t.Errorf("second folder name: got %q, want %q", list.Folders[1].Name, "Action")
	}
}

func TestWriteOmitsEmptyFields(t *testing.T) {
	list := &GameList{
		Games: []Game{
			{
				Path: "./test.gba",
				Name: "Test Game",
				// All other fields left as zero values
			},
		},
	}

	data, err := Write(list)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	xmlStr := string(data)

	// Should contain required fields
	if !contains(xmlStr, "<path>./test.gba</path>") {
		t.Error("output should contain path")
	}
	if !contains(xmlStr, "<name>Test Game</name>") {
		t.Error("output should contain name")
	}

	// Should not contain empty/zero fields
	if contains(xmlStr, "<desc>") {
		t.Error("output should not contain empty desc")
	}
	if contains(xmlStr, "<rating>") {
		t.Error("output should not contain zero rating")
	}
	if contains(xmlStr, "<releasedate>") {
		t.Error("output should not contain zero releasedate")
	}
	if contains(xmlStr, "<players>") {
		t.Error("output should not contain zero players")
	}
	if contains(xmlStr, "<playcount>") {
		t.Error("output should not contain zero playcount")
	}
}

func TestRoundtrip(t *testing.T) {
	original := &GameList{
		Games: []Game{
			{
				Path:        "./test.gba",
				Name:        "Test Game",
				Desc:        "A description",
				Rating:      0.75,
				ReleaseDate: DateTime{Time: time.Date(1995, 12, 25, 0, 0, 0, 0, time.UTC)},
				Developer:   "Dev",
				Publisher:   "Pub",
				Genre:       "Action",
				Players:     4,
			},
		},
		Folders: []Folder{
			{
				Path: "./folder",
				Name: "My Folder",
				Desc: "Folder description",
			},
		},
	}

	// Write to XML
	data, err := Write(original)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Parse back
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify games
	if len(parsed.Games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(parsed.Games))
	}
	g := parsed.Games[0]
	if g.Path != original.Games[0].Path {
		t.Errorf("Path mismatch: got %q, want %q", g.Path, original.Games[0].Path)
	}
	if g.Name != original.Games[0].Name {
		t.Errorf("Name mismatch: got %q, want %q", g.Name, original.Games[0].Name)
	}
	if g.Desc != original.Games[0].Desc {
		t.Errorf("Desc mismatch: got %q, want %q", g.Desc, original.Games[0].Desc)
	}
	if g.Rating != original.Games[0].Rating {
		t.Errorf("Rating mismatch: got %v, want %v", g.Rating, original.Games[0].Rating)
	}
	if !g.ReleaseDate.Equal(original.Games[0].ReleaseDate.Time) {
		t.Errorf("ReleaseDate mismatch: got %v, want %v", g.ReleaseDate.Time, original.Games[0].ReleaseDate.Time)
	}
	if g.Players != original.Games[0].Players {
		t.Errorf("Players mismatch: got %d, want %d", g.Players, original.Games[0].Players)
	}

	// Verify folders
	if len(parsed.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(parsed.Folders))
	}
	f := parsed.Folders[0]
	if f.Path != original.Folders[0].Path {
		t.Errorf("Folder Path mismatch: got %q, want %q", f.Path, original.Folders[0].Path)
	}
	if f.Name != original.Folders[0].Name {
		t.Errorf("Folder Name mismatch: got %q, want %q", f.Name, original.Folders[0].Name)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
