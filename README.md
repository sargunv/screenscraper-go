# screenscraper-go

A Go client library and CLI for Screenscraper, a community platform for retro video game data and media.

## Installation

Install the CLI:

    go install sargunv/screenscraper-go/cmd/screenscraper@latest

Or use the library in your Go project:

    go get sargunv/screenscraper-go

## CLI Usage

```bash
screenscraper --help
```

```
A CLI client for the Screenscraper API to fetch game metadata and media.

Credentials are loaded from environment variables:
  SCREENSCRAPER_DEV_USER     - Developer username
  SCREENSCRAPER_DEV_PASSWORD - Developer password
  SCREENSCRAPER_ID           - User ID (optional)
  SCREENSCRAPER_PASSWORD     - User password (optional)

Usage:
  screenscraper [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download media files
  game        Get game information
  genres      Get list of genres
  help        Help about any command
  infra       Get infrastructure/server information
  languages   Get list of languages
  media-types Get list of media types
  regions     Get list of regions
  search      Search for games by name
  systems     Get list of systems/consoles
  user        Get user information and quotas

Flags:
      --dev-id string          Developer ID (or set SCREENSCRAPER_DEV_USER)
      --dev-password string    Developer password (or set SCREENSCRAPER_DEV_PASSWORD)
  -h, --help                   help for scraper
      --soft-name string       Software name identifier
      --user-id string         User ID (or set SCREENSCRAPER_ID)
      --user-password string   User password (or set SCREENSCRAPER_PASSWORD)

Use "scraper [command] --help" for more information about a command.
```

## Library Usage

    import "sargunv/screenscraper-go/client"

    c := client.NewClient(devID, devPassword, "my-app/1.0", ssID, ssPassword)
    game, err := client.GetGame(client.GetGameParams{GameID: "12345"})

## API Endpoint Implementation Status

### Core Information

- [x] `ssinfraInfos.php` - Infrastructure/server information
- [x] `ssuserInfos.php` - User information and quotas

### Metadata Lists

- [x] `regionsListe.php` - List of regions
- [x] `languesListe.php` - List of languages
- [x] `genresListe.php` - List of genres
- [x] `mediasSystemeListe.php` - List of system media types
- [x] `mediasJeuListe.php` - List of game media types
- [x] `systemesListe.php` - List of systems/consoles
- [ ] `userlevelsListe.php` - List of user levels
- [ ] `nbJoueursListe.php` - List of player counts
- [ ] `supportTypesListe.php` - List of support types
- [ ] `romTypesListe.php` - List of ROM types
- [ ] `famillesListe.php` - List of families
- [ ] `classificationsListe.php` - List of classifications
- [ ] `infosJeuListe.php` - List of game info types
- [ ] `infosRomListe.php` - List of ROM info types

### Game Data

- [x] `jeuRecherche.php` - Search for games by name
- [x] `jeuInfos.php` - Get detailed game information

### Media Downloads

- [x] `mediaJeu.php` - Download game image media
- [x] `mediaSysteme.php` - Download system image media
- [ ] `mediaVideoSysteme.php` - Download system video media
- [ ] `mediaVideoJeu.php` - Download game video media
- [ ] `mediaManuelJeu.php` - Download game manuals (PDF)
- [ ] `mediaGroup.php` - Download group image media (genres, modes, etc.)
- [ ] `mediaCompagnie.php` - Download company image media

### Community Features

- [ ] `botNote.php` - Submit game ratings
- [ ] `botProposition.php` - Submit info/media proposals

## Credentials

You need a Screenscraper developer account. Register at https://www.screenscraper.fr.

User credentials are optional but provide higher rate limits.
