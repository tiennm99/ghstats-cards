# Configurable start-of-week

## Goal
Let users pick which weekday the contribution heatmap rows and productive-weekday bars start on. Default stays Sunday for back-compat with GitHub's own calendar.

## Surface
- CLI: `-start-of-week sunday|monday` (default `sunday`)
- Action input: `start_of_week` (default `sunday`)
- Accept case-insensitive full name; anything else → warn and fall back to Sunday (mirrors existing `-tz` behaviour)

Scope intentionally limited to Sun/Mon — the two real-world cases. Re-open if someone asks for Tue/Wed/etc.

## Code touchpoints
| File | Change |
|------|--------|
| `main.go` | parse `-start-of-week`, set `Profile.WeekStart` |
| `internal/github/model.go` | add `WeekStart time.Weekday` |
| `internal/card/contributions_heatmap.go` | rotate `padToWeekGrid` offset + weekday labels using `WeekStart` |
| `internal/card/productive_weekday.go` | reorder bars and labels using `WeekStart` |
| `action.yml` + `entrypoint.sh` | new input plumbed through |
| `README.md` | document |
| tests | weekday parse + rotation grid |

Streak card is count-based — untouched.

## Phases
1. Wire flag + profile field + action plumbing
2. Heatmap rotation
3. Weekday rotation
4. Tests
5. README

## Success
- `go build ./...` clean
- `go test ./...` passes (existing + new)
- Running with `-start-of-week monday` visibly rotates row 0 of the heatmap to Mon and moves Sun to row 6
- Default behaviour (no flag) identical to current output
