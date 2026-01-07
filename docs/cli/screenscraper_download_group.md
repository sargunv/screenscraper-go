## screenscraper download group

Download group media

### Synopsis

Download group media (genres, modes, families, themes, styles)

```
screenscraper download group [flags]
```

### Examples

```
  # Download genre logo
  screenscraper download group --group-id=1 --media="logo-monochrome" --output=genre.png
```

### Options

```
      --format string       Output format: png or jpg
  -g, --group-id string     Group ID (required)
  -h, --help                help for group
      --max-height string   Maximum height in pixels
      --max-width string    Maximum width in pixels
  -m, --media string        Media identifier (required, e.g. 'logo-monochrome')
  -o, --output string       Output file path (default: stdout)
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

- [screenscraper download](screenscraper_download.md) - Download media files
