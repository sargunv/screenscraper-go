package romident

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/lib/romident/container"
	"github.com/sargunv/rom-tools/lib/romident/format"
)

// IdentifyROM identifies a ROM file, ZIP archive, or folder.
// Returns a ROM struct with file information and hashes.
func IdentifyROM(path string, opts Options) (*ROM, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return identifyFolder(absPath, opts)
	}

	return identifyFile(absPath, info.Size(), opts)
}

// identifyContainer processes any Container (folder, ZIP, etc.) and identifies all files within.
func identifyContainer(c container.Container, containerType ROMType, containerPath string, opts Options) (*ROM, error) {
	entries := c.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("container is empty")
	}

	files := make(Files)
	detector := format.NewDetector()
	var romIdent *GameIdent

	for _, entry := range entries {
		// Open file for identification
		reader, err := c.OpenFileAt(entry.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
		}

		romFile, fileIdent, err := identifySingleReader(reader, entry.Name, detector, opts)
		if err != nil {
			reader.Close()
			return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
		}

		// Add pre-computed CRC32 from container metadata if available and not already calculated
		if entry.CRC32 != 0 {
			hasCRC32 := false
			for _, h := range romFile.Hashes {
				if h.Algorithm == HashCRC32 {
					hasCRC32 = true
					break
				}
			}
			if !hasCRC32 {
				romFile.Hashes = append(romFile.Hashes, NewHash(HashCRC32, fmt.Sprintf("%08x", entry.CRC32), HashSourceZIPMetadata))
			}
		}

		// Collect identification (error if multiple identifications found)
		if fileIdent != nil {
			if romIdent != nil {
				reader.Close()
				return nil, fmt.Errorf("container has multiple game identifications: %s and %s", romIdent.TitleID, fileIdent.TitleID)
			}
			romIdent = fileIdent
		}

		reader.Close()
		files[entry.Name] = *romFile
	}

	return &ROM{
		Path:  containerPath,
		Type:  containerType,
		Files: files,
		Ident: romIdent,
	}, nil
}

func identifyFolder(path string, opts Options) (*ROM, error) {
	c, err := container.NewFolderContainer(path)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return identifyContainer(c, ROMTypeFolder, path, opts)
}

func identifyFile(path string, size int64, opts Options) (*ROM, error) {
	// Open file for format detection
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Detect format
	detector := format.NewDetector()
	detectedFormat, err := detector.Detect(f, size, filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// Handle ZIP specially
	if detectedFormat == format.ZIP {
		return identifyZIP(path, opts)
	}

	// Single file
	romFile, ident, err := identifySingleFile(path, size, detector, opts)
	if err != nil {
		return nil, err
	}

	files := Files{
		filepath.Base(path): *romFile,
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeFile,
		Files: files,
		Ident: ident,
	}, nil
}

func identifyZIP(path string, opts Options) (*ROM, error) {
	handler := container.NewZIPHandler()

	archive, err := handler.Open(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	entries := archive.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("ZIP archive is empty")
	}

	// Slow mode: use full container introspection (decompresses files)
	if opts.HashMode == HashModeSlow {
		return identifyContainer(archive, ROMTypeZIP, path, opts)
	}

	// Fast/default mode: use ZIP metadata only (no decompression)
	files := make(Files)
	detector := format.NewDetector()

	for _, entry := range entries {
		// Use extension-based format detection (no decompression)
		detectedFormat := detector.DetectByExtension(entry.Name)

		hashes := []Hash{}
		if entry.CRC32 != 0 {
			hashes = []Hash{
				NewHash(HashCRC32, fmt.Sprintf("%08x", entry.CRC32), HashSourceZIPMetadata),
			}
		}

		files[entry.Name] = ROMFile{
			Size:   entry.Size,
			Format: formatToRomidentFormat(detectedFormat),
			Hashes: hashes,
		}
	}

	return &ROM{
		Path:  path,
		Type:  ROMTypeZIP,
		Files: files,
		Ident: nil, // No identification in fast/default mode
	}, nil
}

// identifySingleReader identifies a file from a ReaderAtSeekCloser (works for any container).
// Returns the ROMFile, game identification (if any), and an error.
func identifySingleReader(r container.ReaderAtSeekCloser, name string, detector *format.Detector, opts Options) (*ROMFile, *GameIdent, error) {
	size := r.Size()

	// Detect format
	detectedFormat, err := detector.Detect(r, size, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to detect format: %w", err)
	}

	romFile := &ROMFile{
		Size:   size,
		Format: formatToRomidentFormat(detectedFormat),
	}

	var ident *GameIdent

	// For CHD, always extract hashes from header (fast, no decompression)
	if detectedFormat == format.CHD {
		chdInfo, err := format.ParseCHDHeader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse CHD header: %w", err)
		}
		romFile.Hashes = []Hash{
			NewHash(HashSHA1, chdInfo.RawSHA1, HashSourceCHDRaw),
			NewHash(HashSHA1, chdInfo.SHA1, HashSourceCHDCompressed),
		}
		return romFile, ident, nil
	}

	// Extract identification for identifiable formats
	if detectedFormat == format.XISO {
		// Reset reader position for XISO parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyXISO(r, size)
		}
	}
	if detectedFormat == format.XBE {
		// Reset reader position for XBE parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyXBE(r, size)
		}
	}
	if detectedFormat == format.GBA {
		// Reset reader position for GBA parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyGBA(r, size)
		}
	}
	if detectedFormat == format.N64 {
		// Reset reader position for N64 parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyN64(r, size)
		}
	}
	if detectedFormat == format.GB {
		// Reset reader position for GB parsing
		if _, err := r.Seek(0, 0); err == nil {
			ident = identifyGB(r, size)
		}
	}

	// Fast mode: skip calculating hashes for large files, but allow small files
	if opts.HashMode == HashModeFast && size >= FastModeSmallFileThreshold {
		return romFile, ident, nil
	}

	// For other formats, calculate hashes
	// Reset reader position
	if _, err := r.Seek(0, 0); err != nil {
		return nil, nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Wrap ReaderAtSeekCloser as io.Reader for CalculateHashes
	reader := &readerAtWrapper{r: r}
	hashes, err := CalculateHashes(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	romFile.Hashes = hashes

	return romFile, ident, nil
}

func identifySingleFile(path string, size int64, detector *format.Detector, opts Options) (*ROMFile, *GameIdent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileReader := &containerFileReader{file: f, size: info.Size()}
	return identifySingleReader(fileReader, filepath.Base(path), detector, opts)
}

// containerFileReader wraps *os.File to implement ReaderAtSeekCloser.
type containerFileReader struct {
	file *os.File
	size int64
}

func (f *containerFileReader) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

func (f *containerFileReader) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *containerFileReader) Size() int64 {
	return f.size
}

func (f *containerFileReader) Close() error {
	return f.file.Close()
}

// readerAtWrapper wraps ReaderAtSeekCloser to implement io.Reader.
type readerAtWrapper struct {
	r   container.ReaderAtSeekCloser
	pos int64
}

func (w *readerAtWrapper) Read(p []byte) (n int, err error) {
	n, err = w.r.ReadAt(p, w.pos)
	w.pos += int64(n)
	return n, err
}

// identifyXISO extracts game identification from an Xbox XISO.
// Returns nil if identification fails (non-fatal).
func identifyXISO(r io.ReaderAt, size int64) *GameIdent {
	info, err := format.ParseXISO(r, size)
	if err != nil {
		return nil
	}

	version := int(info.Version)
	discNumber := int(info.DiscNumber)

	return &GameIdent{
		Platform:   PlatformXbox,
		TitleID:    fmt.Sprintf("%s-%03d", info.PublisherCode, info.GameNumber),
		Title:      info.Title,
		Regions:    decodeXboxRegions(info.RegionFlags),
		MakerCode:  info.PublisherCode,
		Version:    &version,
		DiscNumber: &discNumber,
		Extra: map[string]string{
			"title_id_hex": info.TitleIDHex,
		},
	}
}

// identifyXBE extracts game identification from an Xbox XBE file.
// Returns nil if identification fails (non-fatal).
func identifyXBE(r io.ReaderAt, size int64) *GameIdent {
	info, err := format.ParseXBE(r, size)
	if err != nil {
		return nil
	}

	version := int(info.Version)
	discNumber := int(info.DiscNumber)

	return &GameIdent{
		Platform:   PlatformXbox,
		TitleID:    fmt.Sprintf("%s-%03d", info.PublisherCode, info.GameNumber),
		Title:      info.Title,
		Regions:    decodeXboxRegions(info.RegionFlags),
		MakerCode:  info.PublisherCode,
		Version:    &version,
		DiscNumber: &discNumber,
		Extra: map[string]string{
			"title_id_hex": info.TitleIDHex,
		},
	}
}

// identifyGBA extracts game identification from a GBA ROM file.
// Returns nil if identification fails (non-fatal).
func identifyGBA(r io.ReaderAt, size int64) *GameIdent {
	info, err := format.ParseGBA(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	return &GameIdent{
		Platform:  PlatformGBA,
		TitleID:   info.GameCode,
		Title:     info.Title,
		Regions:   []Region{decodeGBARegion(info.RegionCode)},
		MakerCode: info.MakerCode,
		Version:   &version,
	}
}

// identifyN64 extracts game identification from an N64 ROM file.
// Returns nil if identification fails (non-fatal).
func identifyN64(r io.ReaderAt, size int64) *GameIdent {
	info, err := format.ParseN64(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	return &GameIdent{
		Platform: PlatformN64,
		TitleID:  info.GameCode,
		Title:    info.Title,
		Regions:  []Region{decodeN64Region(info.RegionCode)},
		Version:  &version,
		Extra: map[string]string{
			"byte_order":    string(info.ByteOrder),
			"category_code": string(info.CategoryCode),
		},
	}
}

// identifyGB extracts game identification from a GB/GBC ROM file.
// Returns nil if identification fails (non-fatal).
func identifyGB(r io.ReaderAt, size int64) *GameIdent {
	info, err := format.ParseGB(r, size)
	if err != nil {
		return nil
	}

	version := info.Version

	// Determine platform based on CGB flag
	var platform Platform
	if info.Platform == format.GBPlatformGBC {
		platform = PlatformGBC
	} else {
		platform = PlatformGB
	}

	// Determine region from destination code
	var region Region
	if info.DestinationCode == 0x00 {
		region = RegionJP
	} else {
		region = RegionWorld // Non-Japanese = worldwide
	}

	extra := map[string]string{
		"licensee": info.LicenseeCode,
	}
	if info.ManufacturerCode != "" {
		extra["manufacturer"] = info.ManufacturerCode
	}
	if info.SGBFlag == format.SGBFlagSupported {
		extra["sgb_support"] = "true"
	}
	extra["cartridge_type"] = fmt.Sprintf("%02X", info.CartridgeType)

	return &GameIdent{
		Platform:  platform,
		Title:     info.Title,
		Regions:   []Region{region},
		MakerCode: info.LicenseeCode,
		Version:   &version,
		Extra:     extra,
	}
}

// decodeXboxRegions converts Xbox region flags to a slice of Region.
func decodeXboxRegions(flags uint32) []Region {
	var regions []Region
	if flags&uint32(format.XboxRegionNA) != 0 {
		regions = append(regions, RegionNA)
	}
	if flags&uint32(format.XboxRegionJapan) != 0 {
		regions = append(regions, RegionJP)
	}
	if flags&uint32(format.XboxRegionEUAU) != 0 {
		regions = append(regions, RegionEU, RegionAU)
	}
	if len(regions) == 0 {
		regions = append(regions, RegionUnknown)
	}
	return regions
}

// decodeGBARegion converts a GBA region code byte to a Region.
func decodeGBARegion(code byte) Region {
	switch code {
	case 'J':
		return RegionJP
	case 'E':
		return RegionUS
	case 'P':
		return RegionEU
	case 'F':
		return RegionFR
	case 'S':
		return RegionES
	case 'D':
		return RegionDE
	case 'I':
		return RegionIT
	default:
		return RegionUnknown
	}
}

// decodeN64Region converts an N64 destination code byte to a Region.
// Based on the N64brew Wiki ROM Header specification.
func decodeN64Region(code byte) Region {
	switch code {
	case 'A':
		return RegionWorld
	case 'B':
		return RegionBR
	case 'C':
		return RegionCN
	case 'D':
		return RegionDE
	case 'E':
		return RegionUS
	case 'F':
		return RegionFR
	case 'G':
		return RegionGatewayNTSC
	case 'H':
		return RegionNL
	case 'I':
		return RegionIT
	case 'J':
		return RegionJP
	case 'K':
		return RegionKR
	case 'L':
		return RegionGatewayPAL
	case 'N':
		return RegionCA
	case 'P', 'X', 'Y', 'Z':
		return RegionEU
	case 'S':
		return RegionES
	case 'U':
		return RegionAU
	case 'W':
		return RegionNordic
	default:
		return RegionUnknown
	}
}

// formatToRomidentFormat converts format.Format to romident.Format
func formatToRomidentFormat(f format.Format) Format {
	switch f {
	case format.CHD:
		return FormatCHD
	case format.XISO:
		return FormatXISO
	case format.XBE:
		return FormatXBE
	case format.ISO9660:
		return FormatISO9660
	case format.ZIP:
		return FormatZIP
	case format.GBA:
		return FormatGBA
	case format.N64:
		return FormatN64
	case format.GB:
		return FormatGB
	default:
		return FormatUnknown
	}
}
