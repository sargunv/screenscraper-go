## rom-tools identify

Identify ROM files and extract metadata

### Synopsis

Extract hashes and game identification data from ROM files.

Supports:

- Hashing files: calculates SHA1, MD5, CRC32 (unless in fast mode).
- CHD files: extracts SHA1 hashes from header (fast, no decompression)
- ZIP archives: extracts CRC32 from metadata (fast, no decompression). If in slow mode, also identifies files within the ZIP.
- Folders: identifies files within.
- XISO and XBE files: identifies Xbox game information from the XBE header.

```
rom-tools identify <file>... [flags]
```

### Options

```
      --fast   Skip hash calculation for large loose files, but calculates for small loose files (<65MiB).
  -h, --help   help for identify
  -j, --json   Output results as JSON Lines (one JSON object per line)
      --slow   Calculate full hashes and identify games inside archives (requires decompression).
```

### SEE ALSO

- [rom-tools](rom-tools.md) - ROM management and metadata tools
