## Overview

- @README.md
- @CONTRIBUTING.md
- @docs/rom-tools.md
- @docs/rom-tools_identify.md
- @docs/rom-tools_scrape.md
- @docs/rom-tools_screenscraper.md

## Tips

- Dev tools are managed with Mise. If your environment is not set up, you'll need to run commands through `mise exec`.
- When adding libraries, use `go get` to get the latest version
- When looking up how to use standard libraries or packages, use `go doc` to read the documentation
- Avoid many redundant tests. Err on the side of too few tests; we'll add more as needed.
