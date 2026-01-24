package chd

import "github.com/sargunv/rom-tools/lib/core"

// Info contains metadata extracted from a CHD file, including identified content.
type Info struct {
	// RawSHA1 is the SHA1 hash of the uncompressed data (from CHD header).
	RawSHA1 string `json:"raw_sha1,omitempty"`
	// SHA1 is the SHA1 hash of the compressed data (from CHD header).
	SHA1 string `json:"sha1,omitempty"`
	// Content is the identified game info from the CHD contents (may be nil).
	Content core.GameInfo `json:"content,omitempty"`
}

// GamePlatform implements core.GameInfo by delegating to Content.
func (i *Info) GamePlatform() core.Platform {
	if i.Content != nil {
		return i.Content.GamePlatform()
	}
	return ""
}

// GameTitle implements core.GameInfo by delegating to Content.
func (i *Info) GameTitle() string {
	if i.Content != nil {
		return i.Content.GameTitle()
	}
	return ""
}

// GameSerial implements core.GameInfo by delegating to Content.
func (i *Info) GameSerial() string {
	if i.Content != nil {
		return i.Content.GameSerial()
	}
	return ""
}
