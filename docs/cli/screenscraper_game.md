## screenscraper game

Get game information

### Synopsis

Retrieves detailed game information including metadata and media URLs.

You can lookup by:

1. ROM hash (CRC/MD5/SHA1) + size + system + name + type (recommended)
2. Game ID (direct lookup)

```
screenscraper game [flags]
```

### Examples

```
  # Lookup by ROM hash
  screenscraper game --crc=50ABC90A --size=749652 --system=1 --rom-type=rom --name="Sonic 2.zip"

  # Lookup by game ID
  screenscraper game --game-id=3
```

### Options

```
      --crc string        ROM CRC32 hash
      --game-id string    Game ID (alternative to hash lookup)
  -h, --help              help for game
      --md5 string        ROM MD5 hash
  -n, --name string       ROM filename
      --rom-type string   ROM type: rom, iso, or folder (default "rom")
      --serial string     Serial number (optional)
      --sha1 string       ROM SHA1 hash
      --size string       ROM size in bytes
  -s, --system string     System ID (required for hash lookup)
```

### Options inherited from parent commands

```
      --dev-id string          Developer ID (or set SCREENSCRAPER_DEV_USER)
      --dev-password string    Developer password (or set SCREENSCRAPER_DEV_PASSWORD)
      --json                   Output results as JSON
      --locale string          Override locale for output (e.g., en, fr, de)
      --user-id string         User ID (or set SCREENSCRAPER_ID)
      --user-password string   User password (or set SCREENSCRAPER_PASSWORD)
```

### SEE ALSO

- [screenscraper](screenscraper.md) - Screenscraper CLI client
