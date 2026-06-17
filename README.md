# Taper

A terminal race tracking and planning app for endurance athletes.

Taper is a Go TUI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Glamour](https://github.com/charmbracelet/glamour). It helps you manage upcoming and past races, keep a journal, work through packing checklists, and store structured race metadata alongside markdown planning docs.

## Features

- **Top-tab navigation** — `Races`, `Calendar`, `Journal`, `Checklist`, `Plan`
- **Training calendar** — month grid with countdown widgets, race-day markers, and 14-day taper highlight
- **Race types** — tag races as road, trail, ultra, or tri (shown in list, calendar, export)
- **Race list** — upcoming and past races sorted by date, with countdown for races within 14 days
- **Structured race metadata** — name, date, distance, location, goal time, status, registration, travel, training flags, post-race fields (stored in `races.json`)
- **Markdown files per race** — `log.md`, `strategy.md`, `packing.md`, `nutrition.md`, `gear.md`, `weather.md`, `crew.md`, `splits.md`
- **External editor** — press `e` to open markdown in `TAPER_EDITOR`, `$EDITOR`, or nano/vim
- **Race packet export** — `x` writes `packet.md`; `X` writes `packet.pdf` via pandoc
- **Glamour read mode** — rendered markdown in a scrollable viewport
- **Interactive checklist** — toggle `- [ ]` / `- [x]` items in `packing.md` with Space
- **Race completion dialog** — press `c` to record finish time, DNF, DNS, lessons learned, and recovery notes

## Install

```bash
git clone <repo> taper
cd taper
make build
./taper
```

Or run directly:

```bash
make run
```

## Data layout

```
~/.config/taper/
├── races.json
└── races/
    └── <race-id>/
        ├── log.md
        ├── strategy.md
        ├── packing.md
        ├── nutrition.md
        ├── gear.md
        ├── weather.md
        ├── crew.md
        └── splits.md
```

`races.json` holds global config and the race list. Each race directory is created automatically when you add a race.

## Keybindings

| Key | Action |
|-----|--------|
| `h` / `l` | Previous / next tab |
| `,` / `.` | Previous / next month (Calendar tab) |
| `j` / `k` | Navigate list, plan docs, or scroll viewport |
| `Enter` | Load highlighted race into memory (opens Journal) |
| `n` | New race |
| `u` | Edit loaded race metadata (multi-section form) |
| `c` | Complete loaded race (outcome dialog) |
| `d` | Delete loaded race (confirm with `y` / `n`) |
| `/` | Filter race list |
| `x` | Export race packet to `packet.md` |
| `X` | Export race packet to `packet.pdf` (requires [pandoc](https://pandoc.org/)) |
| `e` | Edit current markdown file in `$EDITOR` |
| `s` | Toggle strategy preview on Races tab |
| `r` | Toggle read / interactive mode on Checklist |
| `Space` | Toggle packing checklist item |
| `q` | Quit |

### In forms (new / edit race)

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous field |
| `[` / `]` | Previous / next section |
| `Enter` | Save |
| `Esc` | Cancel |

### In completion dialog

| Key | Action |
|-----|--------|
| `Space` | Cycle outcome: completed → dnf → dns |
| `Tab` | Next field |
| `Enter` | Save |
| `Esc` | Cancel |

## What to track

**Structured fields** (in `races.json`, editable via `u`):

- Core: name, date, distance, location, goal time, status
- Registration: entry fee, confirmation #, bib pickup
- Travel: hotel, flights, transport to start
- Training: last long run, peak week, health / taper notes
- Post-race: result time, photos link, lessons learned, recovery plan

**Markdown files** (editable via `e` on the relevant tab):

| File | Purpose |
|------|---------|
| `log.md` | Journal — pre/during/post race notes |
| `packing.md` | Interactive race-day checklist |
| `strategy.md` | Pacing and mental plan |
| `nutrition.md` | Fueling schedule, caffeine timing |
| `gear.md` | Shoes, kit, watch settings |
| `weather.md` | Heat/rain contingency plans |
| `crew.md` | Support contacts and meetup points |
| `splits.md` | Target pace per segment |

## Development

```bash
make test    # run tests
make fmt     # format
make lint    # go vet
make tidy    # go mod tidy
```

## License

MIT
