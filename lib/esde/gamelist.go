// Package esde provides types and utilities for ES-DE (EmulationStation Desktop Edition) integration.
//
// ES-DE uses gamelist.xml files to store game metadata and finds media files
// by naming convention in media subdirectories.
//
// Gamelist.xml specification:
// https://github.com/batocera-linux/batocera-emulationstation/blob/master/GAMELISTS.md
package esde

import (
	"encoding/xml"
	"time"
)

// DateTimeFormat is the format used by ES-DE for dates: YYYYMMDDTHHMMSS
const DateTimeFormat = "20060102T150405"

// DateTime wraps time.Time with ES-DE's YYYYMMDDTHHMMSS format
type DateTime struct {
	time.Time
}

// MarshalXML formats the DateTime as YYYYMMDDTHHMMSS
func (d DateTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d.IsZero() {
		return nil
	}
	return e.EncodeElement(d.Format(DateTimeFormat), start)
}

// UnmarshalXML parses YYYYMMDDTHHMMSS format into DateTime
func (d *DateTime) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// GameList represents an ES-DE gamelist.xml
type GameList struct {
	XMLName xml.Name `xml:"gameList"`
	Games   []Game   `xml:"game"`
	Folders []Folder `xml:"folder"`
}

// Game represents a single game entry in gamelist.xml
type Game struct {
	Path        string   `xml:"path"`
	Name        string   `xml:"name"`
	Desc        string   `xml:"desc,omitempty"`
	Image       string   `xml:"image,omitempty"`
	Thumbnail   string   `xml:"thumbnail,omitempty"`
	Video       string   `xml:"video,omitempty"`
	Rating      float64  `xml:"rating,omitempty"`
	ReleaseDate DateTime `xml:"releasedate,omitempty"`
	Developer   string   `xml:"developer,omitempty"`
	Publisher   string   `xml:"publisher,omitempty"`
	Genre       string   `xml:"genre,omitempty"`
	Players     int      `xml:"players,omitempty"`
	PlayCount   int      `xml:"playcount,omitempty"`
	LastPlayed  DateTime `xml:"lastplayed,omitempty"`
}

// Folder represents a folder entry in gamelist.xml
type Folder struct {
	Path      string `xml:"path"`
	Name      string `xml:"name"`
	Desc      string `xml:"desc,omitempty"`
	Image     string `xml:"image,omitempty"`
	Thumbnail string `xml:"thumbnail,omitempty"`
}

// Parse parses gamelist.xml data into a GameList
func Parse(data []byte) (*GameList, error) {
	var gamelist GameList
	if err := xml.Unmarshal(data, &gamelist); err != nil {
		return nil, err
	}
	return &gamelist, nil
}

// Write serializes a GameList to XML with proper formatting
func Write(list *GameList) ([]byte, error) {
	data, err := xml.MarshalIndent(list, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), append(data, '\n')...), nil
}
