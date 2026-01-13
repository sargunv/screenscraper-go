package romident

import "io"

// Magic bytes and offsets for container/disc formats (non-game formats).
// Game format verification is delegated to the format registry identifiers.
var (
	// CHD v5: magic at start
	chdMagic  = []byte("MComprHD")
	chdOffset = int64(0)

	// ISO9660: magic at offset 0x8001 (sector 16 + 1 byte for type code)
	iso9660Magic  = []byte("CD001")
	iso9660Offset = int64(0x8001)

	// ZIP: magic at start
	zipMagic  = []byte{0x50, 0x4B, 0x03, 0x04}
	zipOffset = int64(0)
)

// Detector can detect the format of a file.
type Detector struct{}

// NewDetector creates a new format detector.
func NewDetector() *Detector {
	return &Detector{}
}

// CandidatesByExtension returns possible formats based on file extension.
// Returns nil for generic/unknown extensions (we don't do magic-only detection).
// This delegates to the format registry for most lookups.
func (d *Detector) CandidatesByExtension(filename string) []Format {
	entries := FormatsByExtension(filename)
	if len(entries) == 0 {
		return nil
	}

	formats := make([]Format, len(entries))
	for i, entry := range entries {
		formats[i] = entry.Format
	}
	return formats
}

// ReaderAtSeeker combines io.ReaderAt and io.Seeker for random access reading.
type ReaderAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// VerifyFormat checks if a file matches a specific format.
// For container/disc formats (CHD, ZIP, ISO9660), uses magic bytes.
// For game formats, uses the format registry's identifier.
func (d *Detector) VerifyFormat(r io.ReaderAt, size int64, format Format) bool {
	// Container/disc formats: use magic byte verification
	switch format {
	case FormatZIP:
		return checkMagic(r, size, zipOffset, zipMagic)
	case FormatCHD:
		return checkMagic(r, size, chdOffset, chdMagic)
	case FormatISO9660:
		return checkMagic(r, size, iso9660Offset, iso9660Magic)
	}

	// Game formats: find the identify function in the registry and try to identify
	for _, entry := range registry {
		if entry.Format == format && entry.Identify != nil {
			_, err := entry.Identify(r, size)
			return err == nil
		}
	}

	return false
}

// checkMagic is a helper to verify magic bytes at a specific offset.
func checkMagic(r io.ReaderAt, size int64, offset int64, magic []byte) bool {
	if size < offset+int64(len(magic)) {
		return false
	}
	buf := make([]byte, len(magic))
	if _, err := r.ReadAt(buf, offset); err != nil {
		return false
	}
	return bytesEqual(buf, magic)
}

// Detect identifies the format using extension to narrow candidates, then verifies.
// Returns Unknown for generic extensions (like .bin) or if verification fails.
// For game formats, this delegates to the format registry's identifiers.
func (d *Detector) Detect(r ReaderAtSeeker, size int64, filename string) (Format, error) {
	candidates := d.CandidatesByExtension(filename)
	if len(candidates) == 0 {
		// Generic or unknown extension: no identification
		return FormatUnknown, nil
	}

	// Try each candidate and return the first that verifies
	for _, candidate := range candidates {
		if d.VerifyFormat(r, size, candidate) {
			return candidate, nil
		}
	}

	// Extension suggested format(s) but none verified
	return FormatUnknown, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
