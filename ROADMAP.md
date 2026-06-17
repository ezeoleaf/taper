# Roadmap

## Done

- [x] Bubble Tea TUI with aerobix-style horizontal tabs
- [x] `races.json` + per-race markdown directories under `~/.config/taper/`
- [x] Race list (upcoming / past) with auto-selection and detail panel
- [x] Journal tab — Glamour-rendered `log.md`
- [x] Checklist tab — interactive `packing.md` checkbox toggling
- [x] Plan tab — nutrition, gear, weather, crew, splits, strategy markdown docs
- [x] External editor via `tea.ExecProcess` + `$EDITOR`
- [x] New / edit race forms with multi-section metadata
- [x] Race completion dialog (`c`) — completed / DNF / DNS + post-race fields
- [x] Structured fields: registration, travel, training block, health, post-race
- [x] Markdown caching + Glamour renderer reuse (tab-switch performance)

## Near term

- [x] Delete race (`d` with confirmation)
- [x] Search / filter race list (`/` on Races tab)
- [x] Export race packet to Markdown (`x`) or PDF (`X` via pandoc)
- [x] `TAPER_EDITOR` env override for default editor
- [x] Auto-migrate / backfill new markdown files for existing installs

## Mid term

- [x] Training calendar view — countdown widgets, taper week highlight (`Calendar` tab)
- [ ] Link to external results (Strava, UltraSignup, Athlinks)
- [ ] Import race from iCal / URL
- [ ] Repeat annual races (copy previous year's plan)
- [x] Tags / race type (road, trail, ultra, tri)

## Longer term

- [ ] Optional sync (git-backed config dir or cloud)
- [ ] Web companion read-only dashboard
- [ ] Integration with aerobix training load for taper guidance
- [ ] Weather API prefetch into `weather.md`

### Design note

Structured fields in `races.json` are for things you want to scan quickly in the race list and detail panel. Long-form planning content lives in markdown so you can edit it freely in your editor of choice. New doc types should default to markdown unless they need to appear in list sorting or completion flows.
