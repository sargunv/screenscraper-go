# Contributing

## Prerequisites

This project uses Mise for managing development tools and dependencies. Install Mise if you haven't already:

https://mise.jdx.dev/

## Setup

1. Clone the repository
2. Run `mise setup` to install required tools (Go, hk, pkl)
3. Create a `.env` file for API credentials (see `.env.example` for reference)

## Development

Build the project:

    mise build

Run tests against the live API:

    mise test

Run the CLI:

    mise screenscraper [args...]

## Documentation

- Screenscraper API documentation is translated from French to English in `api/SCREENSCRAPER_API.md`

## Organization

- `client/` directory contains an implementation file and test file per API endpoint
- `internal/cli/` directory contains a command file per CLI command
- `api/` directory contains the API documentation translated from French to English
