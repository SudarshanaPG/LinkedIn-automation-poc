# LinkedIn Automation POC (Educational Only)

This repository contains a Go-based proof-of-concept for advanced browser automation and stealth techniques using the Rod library. It is **strictly for technical evaluation and education**.

## Critical Disclaimer

- Automating LinkedIn violates their Terms of Service and can lead to permanent account bans or legal action.
- This project is **not** for production use. Use only in controlled, educational contexts.
- Do not run this against real accounts or production data.

## Features

Core automation:
- Login using credentials from environment variables
- Cookie persistence and session reuse
- Search people by companies list (company-only) or by title/company/location/keywords
- Optional `search.leads_file` mode to bypass search scraping
- Pagination handling and duplicate suppression
- Connection requests with personalized notes
- Accepted-connection detection and follow-up messaging
- JSON state persistence for sent requests and messages

Stealth techniques (10+):
- Human-like mouse movement using bezier curves, overshoot, micro-corrections
- Randomized timing and think-time delays
- Browser fingerprint masking (`navigator.webdriver`, languages, plugins, WebGL)
- User agent overrides and viewport randomization
- Random scrolling with backtracking
- Realistic typing with typos and backspaces
- Mouse hovering and cursor wandering
- Activity scheduling with business-hour windows and breaks
- Rate limiting and hourly/daily throttling
- Action cadence variation between steps
- Retry with exponential backoff on transient failures (navigation/element lookup/click)

## Project Layout

- `cmd/linkedin-poc/main.go` entrypoint
- `internal/auth` login, cookies, checkpoint detection
- `internal/search` people search and parsing
- `internal/connect` connection requests and notes
- `internal/messaging` accepted detection and follow-ups
- `internal/stealth` anti-detection behaviors
- `internal/limits` rate limiting
- `internal/storage` JSON persistence
- `internal/browser` Rod setup and fingerprint injection

## Setup

Prerequisites:
- Go 1.21+
- Chrome/Edge/Chromium installed locally (or set `browser.executable_path` / `LINKEDIN_BROWSER_PATH`)

1) Copy environment template and fill in credentials:
```bash
copy .env.example .env
```

2) Optionally copy config template:
```bash
copy config.yaml.example config.yaml
```

3) Fetch dependencies:
```bash
go mod tidy
```

4) Run:
```bash
go run ./cmd/linkedin-poc -config config.yaml
```

### Dry Run (Recommended)

Runs a local demo page to showcase stealth behaviors (mouse movement, typing rhythm, scrolling) **without** logging into LinkedIn or clicking Connect/Message:

```bash
go run ./cmd/linkedin-poc -config config.yaml -dry-run
```

If you see an error like “can't find a browser binary”, set one of:
- `LINKEDIN_BROWSER_PATH` in `.env` (e.g. `C:\Program Files\Google\Chrome\Application\chrome.exe`)
- `browser.executable_path` in `config.yaml`

## Mock End-to-End Demo (No LinkedIn)

If you need to show the full login → search → connect → message flow without using LinkedIn, run the included mock site locally:

1) Start the mock server:
```bash
go run ./cmd/mock-linkedin -addr :7777
```

2) In a second terminal, run the automation against the mock config:
```bash
go run ./cmd/linkedin-poc -config config.mock.yaml.example
```

This produces `data/mock-state.json` and `data/mock-cookies.json` and exercises the real automation modules against local pages.

You can also demo the connect/message flow without scraping search results by using a pre-defined leads file:
```bash
go run ./cmd/linkedin-poc -config config.mock.leads.yaml.example
```

## Configuration

- `.env` overrides credentials and high-level runtime flags.
- `config.yaml` controls search filters, limits, stealth behavior, and storage paths.
- If `config.yaml` is missing, defaults are applied automatically.
  - `search.companies` + `search.company_only=true` runs people-search per company, with `search.max_pages` applied per company.
  - `search.leads_file` loads leads from a text file (one URL per line) and skips search scraping entirely.

## Demo Video

- Demo 1: https://drive.google.com/file/d/1GJbep-1qyapiNIHdj5bEszfhp2L6sn2z/view?usp=drive_link
- Demo 2: https://drive.google.com/file/d/1WMcp_Z0A_cLv3wBDcVqaFtvbegrbVsTk/view?usp=drive_link

## Notes

- Selectors used in search and messaging may require updates if LinkedIn changes DOM structure.
- The default schedule enforces business-hour windows but does not hard-block by default (`schedule.enforce=false`).
- This is a proof-of-concept focused on stealth and architecture, not production reliability.
