## rom-tools scrape

Scrape metadata for ROM collections

### Synopsis

Batch scrape metadata and media for ROM files.

Scans the input (DAT file or ROM directory), identifies games using hashes, fetches metadata from Screenscraper, downloads media files, and generates output in the specified format(s).

Example:

# Scrape from DAT file to ES-DE format

rom-tools scrape --system megadrive --dat megadrive.dat \
 --esde-gamelist ./roms/megadrive/gamelist.xml \
 --esde-media ./roms/megadrive/media

# Scrape with custom media types and regions

rom-tools scrape --system gba --dat gba.dat \
 --esde-gamelist ./gba/gamelist.xml \
 --esde-media ./gba/media \
 --media screenshots,covers,3dboxes,marquees,videos \
 --regions jp,us,eu

# Dry run to see what would be scraped

rom-tools scrape --system snes --dat snes.dat --dry-run

Common systems: megadrive, gba, snes, nes, psx, ps2, dreamcast, n64, nds, gb, gbc. Use 'rom-tools screenscraper list systems' to see all available systems.

```
rom-tools scrape [flags]
```

### Options

```
      --cache-age duration      Maximum cache age (default 30 days) (default 720h0m0s)
      --cache-only              Only use cached data, no API calls
  -d, --dat string              Path to DAT file (Logiqx XML format)
      --dry-run                 Parse input and show what would be scraped
      --esde-gamelist string    Path for ES-DE gamelist.xml
      --esde-media string       Path for ES-DE media folder
      --fast                    Skip hash calculation for large files
      --filter string           Filter expression for which games to scrape (e.g., 'missing.metadata', 'missing.covers or missing.videos') (default "true")
  -h, --help                    help for scrape
      --http-timeout duration   HTTP request timeout (e.g., 30s, 2m, 5m) (default 5m0s)
  -i, --input string            Path to ROM directory (not yet implemented)
  -j, --json                    Output final results as JSON
  -m, --media strings           Media types to download: screenshots,titlescreens,covers,3dboxes,marquees,fanart,videos,physicalmedia,backcovers (default [screenshots,covers,marquees])
      --no-cache                Don't read from cache (still writes to cache)
      --overwrite               Overwrite existing media files and gamelist entries
  -r, --regions strings         Preferred regions in order (default [us,eu,jp])
      --slow                    Calculate full hashes for archives
  -s, --system string           System name or ID (e.g., megadrive, gba, snes, psx)
      --threads int             Max concurrent API requests (0 = use account limit)
```

### SEE ALSO

- [rom-tools](rom-tools.md) - ROM management and metadata tools
