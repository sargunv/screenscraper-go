package datfile

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// DumpStatus represents the verification status of a ROM or Disk dump
type DumpStatus string

const (
	DumpStatusUnspecified DumpStatus = ""     // zero value when unset
	DumpStatusGood        DumpStatus = "good" // DTD default
	DumpStatusBadDump     DumpStatus = "baddump"
	DumpStatusNoDump      DumpStatus = "nodump"
	DumpStatusVerified    DumpStatus = "verified"
)

// MergeMode represents merging behavior for ROM sets
type MergeMode string

const (
	MergeModeUnspecified MergeMode = "" // zero value when unset
	MergeModeNone        MergeMode = "none"
	MergeModeSplit       MergeMode = "split" // DTD default for forcemerging, rommode, biosmode
	MergeModeFull        MergeMode = "full"
	MergeModeMerged      MergeMode = "merged" // DTD default for samplemode
	MergeModeUnmerged    MergeMode = "unmerged"
)

// NoDumpMode represents how to handle ROMs with no dump available
type NoDumpMode string

const (
	NoDumpModeUnspecified NoDumpMode = ""         // zero value when unset
	NoDumpModeObsolete    NoDumpMode = "obsolete" // DTD default
	NoDumpModeRequired    NoDumpMode = "required"
	NoDumpModeIgnore      NoDumpMode = "ignore"
)

// PackingMode represents how ROMs should be packed
type PackingMode string

const (
	PackingModeUnspecified PackingMode = ""    // zero value when unset
	PackingModeZip         PackingMode = "zip" // DTD default
	PackingModeUnzip       PackingMode = "unzip"
)

// Datafile represents a parsed DAT file
type Datafile struct {
	Header Header
	Games  []Game
}

// Header contains metadata about the DAT file
type Header struct {
	ID          *int // No-Intro only
	Name        string
	Description string
	Category    string
	Version     string
	Date        string
	Author      string
	Email       string
	Homepage    string
	URL         string
	Comment     string
	Subset      string // No-Intro only
	ClrMamePro  *ClrMamePro
	RomCenter   *RomCenter
}

func (h *Header) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type rawRomCenter struct {
		Plugin         string `xml:"plugin,attr"`
		RomMode        string `xml:"rommode,attr"`
		BiosMode       string `xml:"biosmode,attr"`
		SampleMode     string `xml:"samplemode,attr"`
		LockRomMode    string `xml:"lockrommode,attr"`
		LockBiosMode   string `xml:"lockbiosmode,attr"`
		LockSampleMode string `xml:"locksamplemode,attr"`
	}
	type rawHeader struct {
		ID          string        `xml:"id"`
		Name        string        `xml:"name"`
		Description string        `xml:"description"`
		Category    string        `xml:"category"`
		Version     string        `xml:"version"`
		Date        string        `xml:"date"`
		Author      string        `xml:"author"`
		Email       string        `xml:"email"`
		Homepage    string        `xml:"homepage"`
		URL         string        `xml:"url"`
		Comment     string        `xml:"comment"`
		Subset      string        `xml:"subset"`
		ClrMamePro  *ClrMamePro   `xml:"clrmamepro"`
		RomCenter   *rawRomCenter `xml:"romcenter"`
	}
	var raw rawHeader
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	h.Name = raw.Name
	h.Description = raw.Description
	h.Category = raw.Category
	h.Version = raw.Version
	h.Date = raw.Date
	h.Author = raw.Author
	h.Email = raw.Email
	h.Homepage = raw.Homepage
	h.URL = raw.URL
	h.Comment = raw.Comment
	h.Subset = raw.Subset
	h.ClrMamePro = raw.ClrMamePro

	if raw.ID != "" {
		if id, err := strconv.Atoi(raw.ID); err == nil {
			h.ID = &id
		}
	}

	if raw.RomCenter != nil {
		h.RomCenter = &RomCenter{
			Plugin:         raw.RomCenter.Plugin,
			RomMode:        MergeMode(raw.RomCenter.RomMode),
			BiosMode:       MergeMode(raw.RomCenter.BiosMode),
			SampleMode:     MergeMode(raw.RomCenter.SampleMode),
			LockRomMode:    parseBool(raw.RomCenter.LockRomMode),
			LockBiosMode:   parseBool(raw.RomCenter.LockBiosMode),
			LockSampleMode: parseBool(raw.RomCenter.LockSampleMode),
		}
	}

	return nil
}

// ClrMamePro contains ClrMamePro-specific options
type ClrMamePro struct {
	Header       string      `xml:"header,attr"`
	ForceMerging MergeMode   `xml:"forcemerging,attr"`
	ForceNoDump  NoDumpMode  `xml:"forcenodump,attr"`
	ForcePacking PackingMode `xml:"forcepacking,attr"`
}

// RomCenter contains RomCenter-specific options
type RomCenter struct {
	Plugin         string
	RomMode        MergeMode
	BiosMode       MergeMode
	SampleMode     MergeMode
	LockRomMode    bool
	LockBiosMode   bool
	LockSampleMode bool
}

// Game represents a game entry in the DAT (also called "machine" in some formats)
type Game struct {
	Name       string
	SourceFile string
	IsBIOS     bool
	CloneOf    string
	RomOf      string
	SampleOf   string
	Board      string
	RebuildTo  string
	ID         string // No-Intro only
	CloneOfID  string // No-Intro only

	Comments     []string
	Description  string
	Year         string
	Manufacturer string
	Categories   []string // No-Intro only

	Releases []Release
	BIOSSets []BIOSSet
	ROMs     []ROM
	Disks    []Disk
	Samples  []Sample
	Archives []Archive
}

func (g *Game) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type rawGame struct {
		Name       string `xml:"name,attr"`
		SourceFile string `xml:"sourcefile,attr"`
		IsBIOS     string `xml:"isbios,attr"`
		CloneOf    string `xml:"cloneof,attr"`
		RomOf      string `xml:"romof,attr"`
		SampleOf   string `xml:"sampleof,attr"`
		Board      string `xml:"board,attr"`
		RebuildTo  string `xml:"rebuildto,attr"`
		ID         string `xml:"id,attr"`
		CloneOfID  string `xml:"cloneofid,attr"`

		Comments     []string  `xml:"comment"`
		Description  string    `xml:"description"`
		Year         string    `xml:"year"`
		Manufacturer string    `xml:"manufacturer"`
		Categories   []string  `xml:"category"`
		Releases     []Release `xml:"release"`
		BIOSSets     []BIOSSet `xml:"biosset"`
		ROMs         []ROM     `xml:"rom"`
		Disks        []Disk    `xml:"disk"`
		Samples      []Sample  `xml:"sample"`
		Archives     []Archive `xml:"archive"`
	}
	var raw rawGame
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	g.Name = raw.Name
	g.SourceFile = raw.SourceFile
	g.IsBIOS = parseBool(raw.IsBIOS)
	g.CloneOf = raw.CloneOf
	g.RomOf = raw.RomOf
	g.SampleOf = raw.SampleOf
	g.Board = raw.Board
	g.RebuildTo = raw.RebuildTo
	g.ID = raw.ID
	g.CloneOfID = raw.CloneOfID
	g.Comments = raw.Comments
	g.Description = raw.Description
	g.Year = raw.Year
	g.Manufacturer = raw.Manufacturer
	g.Categories = raw.Categories
	g.Releases = raw.Releases
	g.BIOSSets = raw.BIOSSets
	g.ROMs = raw.ROMs
	g.Disks = raw.Disks
	g.Samples = raw.Samples
	g.Archives = raw.Archives

	return nil
}

// ROM represents a ROM file entry
type ROM struct {
	Name   string
	Size   int64
	CRC    string
	SHA1   string
	MD5    string
	SHA256 string // No-Intro only
	Merge  string
	Status DumpStatus
	Date   string
	Serial string // No-Intro only
	Header string // No-Intro only
}

func (r *ROM) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type rawROM struct {
		Name   string `xml:"name,attr"`
		Size   string `xml:"size,attr"`
		CRC    string `xml:"crc,attr"`
		SHA1   string `xml:"sha1,attr"`
		MD5    string `xml:"md5,attr"`
		SHA256 string `xml:"sha256,attr"`
		Merge  string `xml:"merge,attr"`
		Status string `xml:"status,attr"`
		Date   string `xml:"date,attr"`
		Serial string `xml:"serial,attr"`
		Header string `xml:"header,attr"`
	}
	var raw rawROM
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	r.Name = raw.Name
	r.Size, _ = strconv.ParseInt(raw.Size, 10, 64)
	r.CRC = raw.CRC
	r.SHA1 = raw.SHA1
	r.MD5 = raw.MD5
	r.SHA256 = raw.SHA256
	r.Merge = raw.Merge
	r.Status = DumpStatus(raw.Status)
	r.Date = raw.Date
	r.Serial = raw.Serial
	r.Header = raw.Header

	return nil
}

// Release represents a release entry
type Release struct {
	Name     string
	Region   string
	Language string
	Date     string
	Default  bool
}

func (r *Release) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type rawRelease struct {
		Name     string `xml:"name,attr"`
		Region   string `xml:"region,attr"`
		Language string `xml:"language,attr"`
		Date     string `xml:"date,attr"`
		Default  string `xml:"default,attr"`
	}
	var raw rawRelease
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	r.Name = raw.Name
	r.Region = raw.Region
	r.Language = raw.Language
	r.Date = raw.Date
	r.Default = parseBool(raw.Default)

	return nil
}

// Disk represents a disk entry (for CD-based systems)
type Disk struct {
	Name   string     `xml:"name,attr"`
	SHA1   string     `xml:"sha1,attr"`
	MD5    string     `xml:"md5,attr"`
	Merge  string     `xml:"merge,attr"`
	Status DumpStatus `xml:"status,attr"`
}

// BIOSSet represents a BIOS set entry
type BIOSSet struct {
	Name        string
	Description string
	Default     bool
}

func (b *BIOSSet) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type rawBIOSSet struct {
		Name        string `xml:"name,attr"`
		Description string `xml:"description,attr"`
		Default     string `xml:"default,attr"`
	}
	var raw rawBIOSSet
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	b.Name = raw.Name
	b.Description = raw.Description
	b.Default = parseBool(raw.Default)

	return nil
}

// Sample represents a sample entry
type Sample struct {
	Name string `xml:"name,attr"`
}

// Archive represents an archive entry
type Archive struct {
	Name string `xml:"name,attr"`
}

// Parse reads and parses a DAT file (Logiqx XML format)
func Parse(path string) (*Datafile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DAT file: %w", err)
	}
	defer f.Close()

	return ParseReader(f)
}

// ParseReader parses a DAT file from a reader
func ParseReader(r io.Reader) (*Datafile, error) {
	// xmlDatafile is used only for top-level parsing to handle both <game> and <machine> elements
	type xmlDatafile struct {
		XMLName  xml.Name `xml:"datafile"`
		Header   Header   `xml:"header"`
		Games    []Game   `xml:"game"`
		Machines []Game   `xml:"machine"`
	}

	var xmlFile xmlDatafile
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&xmlFile); err != nil {
		return nil, fmt.Errorf("failed to parse DAT file: %w", err)
	}

	file := &Datafile{
		Header: xmlFile.Header,
		Games:  make([]Game, 0, len(xmlFile.Games)+len(xmlFile.Machines)),
	}
	file.Games = append(file.Games, xmlFile.Games...)
	file.Games = append(file.Games, xmlFile.Machines...)

	return file, nil
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "yes" || s == "true" || s == "1"
}
