package core

// HashType identifies a specific hash algorithm and source.
type HashType string

const (
	// Calculated hash types (computed from file content)
	HashSHA1  HashType = "sha1"
	HashMD5   HashType = "md5"
	HashCRC32 HashType = "crc32"

	// Container metadata hash types (extracted from archive headers)
	HashZipCRC32 HashType = "zip-crc32"

	// CHD hash types (extracted from CHD file headers)
	HashCHDUncompressedSHA1 HashType = "chd-uncompressed-sha1"
	HashCHDCompressedSHA1   HashType = "chd-compressed-sha1"
)

// Hashes maps hash type to hex-encoded value.
type Hashes map[HashType]string
