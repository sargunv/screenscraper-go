package identify

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sargunv/rom-tools/internal/container/folder"
	"github.com/sargunv/rom-tools/internal/container/zip"
	"github.com/sargunv/rom-tools/internal/util"
	"github.com/sargunv/rom-tools/lib/chd"
	"github.com/sargunv/rom-tools/lib/core"
)

// ZIP magic bytes
var zipMagic = []byte{0x50, 0x4B, 0x03, 0x04}

// Identify identifies a ROM file, ZIP archive, or folder.
// Returns a Result with identified items and their hashes.
func Identify(path string, opts Options) (*Result, error) {
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

// identifyFile handles a single file (may be a container like ZIP).
func identifyFile(path string, size int64, opts Options) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Check if it's a ZIP file (special handling required)
	if isZIP(f, size) {
		return identifyZIP(path, opts)
	}

	// Single file - identify it
	item, err := identifyReader(f, size, filepath.Base(path), opts)
	if err != nil {
		return nil, err
	}

	return &Result{
		Path:  path,
		Items: []Item{*item},
	}, nil
}

// identifyFolder handles a directory containing ROM files.
func identifyFolder(path string, opts Options) (*Result, error) {
	c, err := folder.NewFolderContainer(path)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	entries := c.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("folder is empty")
	}

	items := make([]Item, 0, len(entries))

	for _, entry := range entries {
		reader, size, err := c.OpenFileAt(entry.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
		}

		item, err := identifyReader(reader, size, entry.Name, opts)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
		}

		items = append(items, *item)
	}

	return &Result{
		Path:  path,
		Items: items,
	}, nil
}

// identifyZIP handles a ZIP archive.
func identifyZIP(path string, opts Options) (*Result, error) {
	handler := zip.NewZIPHandler()

	archive, err := handler.Open(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	entries := archive.Entries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("ZIP archive is empty")
	}

	items := make([]Item, 0, len(entries))

	for _, entry := range entries {
		if opts.HashMode == HashModeSlow {
			// Slow mode: decompress and fully identify
			reader, size, err := archive.OpenFileAt(entry.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to open %s: %w", entry.Name, err)
			}

			item, err := identifyReader(reader, size, entry.Name, opts)
			reader.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to identify %s: %w", entry.Name, err)
			}

			items = append(items, *item)
		} else {
			// Fast/default mode: use ZIP metadata only (no decompression)
			hashes := make(Hashes)
			if entry.CRC32 != 0 {
				hashes[HashZipCRC32] = fmt.Sprintf("%08x", entry.CRC32)
			}

			items = append(items, Item{
				Name:   entry.Name,
				Size:   entry.Size,
				Hashes: hashes,
				Game:   nil, // No identification in fast mode
			})
		}
	}

	return &Result{
		Path:  path,
		Items: items,
	}, nil
}

// identifyReader identifies a single file from a reader.
// Returns an Item with hashes and game info.
// Accepts either *os.File or util.RandomAccessReader.
func identifyReader(r util.RandomAccessReader, size int64, name string, opts Options) (*Item, error) {
	// Try to identify game
	game := identifyGame(r, size, name)

	item := &Item{
		Name: name,
		Size: size,
		Game: game,
	}

	// Handle CHD: extract hashes from the parsed info
	if chdInfo, ok := game.(*chd.Info); ok {
		item.Hashes = Hashes{
			HashCHDUncompressedSHA1: chdInfo.RawSHA1,
			HashCHDCompressedSHA1:   chdInfo.SHA1,
		}
		return item, nil
	}

	// Fast mode: skip hashes for large files
	if opts.HashMode == HashModeFast && size >= fastModeSmallFileThreshold {
		return item, nil
	}

	// Calculate hashes
	hashes, err := calculateHashes(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hashes: %w", err)
	}

	item.Hashes = hashes
	return item, nil
}

// identifyGame tries to identify the game from a reader.
// Returns the game info (nil if not identifiable).
func identifyGame(r util.RandomAccessReader, size int64, name string) core.GameInfo {
	// Get candidate parsers by extension
	parsers := identifyByExtension(name)
	if len(parsers) == 0 {
		return nil
	}

	// Try each parser
	for _, parser := range parsers {
		// Reset reader position
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			continue
		}

		// Try to identify using the parser
		game, err := parser(r, size)
		if err == nil && game != nil {
			return game
		}
	}

	return nil
}

// isZIP checks if a file is a ZIP archive by checking magic bytes.
func isZIP(r io.ReaderAt, size int64) bool {
	if size < int64(len(zipMagic)) {
		return false
	}
	buf := make([]byte, len(zipMagic))
	if _, err := r.ReadAt(buf, 0); err != nil {
		return false
	}
	return bytes.Equal(buf, zipMagic)
}
