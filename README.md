# rom-tools

A CLI to scrape ROM metadata and media from Screenscraper, a community platform for retro video data.

## CLI

Install the CLI:

    go install github.com/sargunv/rom-tools/cmd/rom-tools

See the [CLI documentation](./docs/rom-tools.md) for complete usage information.

## Libraries

### ScreenScraper API client

To use the ScreenScraper API client in your Go project:

    go get github.com/sargunv/rom-tools/lib/screenscraper

## Test Data

ROM files in `**/testdata/` are sourced from:

- [XboxDev/cromwell](https://github.com/XboxDev/cromwell) (LGPL-2.1)
- [Zophar's Domain PD ROMs](https://www.zophar.net/pdroms/) (public domain)

These files are used as sample data for automated tests.
