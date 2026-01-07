## screenscraper

Screenscraper CLI client

### Synopsis

A CLI client for the Screenscraper API to fetch game metadata and media.

Credentials are loaded from environment variables:

- SCREENSCRAPER_DEV_USER - Developer username
- SCREENSCRAPER_DEV_PASSWORD - Developer password
- SCREENSCRAPER_ID - User ID (optional)
- SCREENSCRAPER_PASSWORD - User password (optional)

### Options

```
      --dev-id string          Developer ID (or set SCREENSCRAPER_DEV_USER)
      --dev-password string    Developer password (or set SCREENSCRAPER_DEV_PASSWORD)
  -h, --help                   help for screenscraper
      --json                   Output results as JSON
      --locale string          Override locale for output (e.g., en, fr, de)
      --user-id string         User ID (or set SCREENSCRAPER_ID)
      --user-password string   User password (or set SCREENSCRAPER_PASSWORD)
```

### SEE ALSO

- [screenscraper download](screenscraper_download.md) - Download media files
- [screenscraper game](screenscraper_game.md) - Get game information
- [screenscraper infra](screenscraper_infra.md) - Get infrastructure/server information
- [screenscraper list](screenscraper_list.md) - List metadata and reference data
- [screenscraper search](screenscraper_search.md) - Search for games by name
- [screenscraper user](screenscraper_user.md) - Get user information and quotas
