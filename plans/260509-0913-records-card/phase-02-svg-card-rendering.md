---
phase: 2
title: "SVG card rendering"
status: completed
priority: P1
effort: "1.5h"
dependencies: [1]
---

# Phase 2: SVG card rendering

## Overview

Render the 6 records as a `records.svg` card mirroring `stats.svg`'s row layout (icon + label + right-aligned accent value). Register the card in `allCards` so `RenderAll` picks it up.

## Requirements

**Functional**
- File output: `records.svg` (340×200, matches all other cards)
- Title: `Records (all time)`
- 6 rows in this order:
  1. ⚡ **Best day** — `{count} on YYYY-MM-DD`
  2. 🔥 **Best month** — `{count} in MMM YYYY` (e.g. `612 in Mar 2026`)
  3. 🌱 **First contribution** — `YYYY-MM-DD`
  4. 📆 **Active days** — `{formatInt(count)}`
  5. ⏳ **On GitHub** — `{years.1f} years`
  6. 🌐 **Languages used** — `{count}`
- Empty-data fallback: render the card with rows showing `—` (em-dash) for missing values, **never panic**

**Non-functional**
- Reuse `header()` / `footer` helpers from `internal/card/svg.go`
- Reuse `escapeXML()` and `formatInt()` from existing helpers
- Octicon paths live in `icons.go` (1-3 new icons may be needed)

## Architecture

```
internal/card/records.go
├── recordsCard struct{}
├── (recordsCard) Filename() string  → "records.svg"
├── (recordsCard) SVG(p, t)          → builds 6 statRow-style entries, emits SVG
└── helpers from Phase 1
```

Register in `internal/card/card.go`:
```go
var allCards = []Card{
    ...
    contributionsByYearCard{},
    recordsCard{},   // new — appended last
}
```

**Icon strategy (minimize new octicon paths):**
- Best day → `iconCommit` (existing, fits "activity peak")
- Best month → `iconStar` (existing, fits "highlight")
- First contribution → **new** `iconCalendar` octicon
- Active days → `iconRepos` (existing, decent fallback) OR **new** `iconHistory`
- On GitHub → `iconClock` (existing)
- Languages used → **new** `iconGlobe` octicon

Net: **2-3 new octicon paths** added to `icons.go` from primer/octicons. Final icon set decided during implementation — fewer-new is fine if reuse looks clean.

## Related Code Files

- Create: `internal/card/records.go`
- Modify: `internal/card/card.go` (register `recordsCard{}` in `allCards`)
- Modify: `internal/card/icons.go` (add 2-3 octicon paths)
- Read for context: `internal/card/stats.go`, `internal/card/svg.go`

## Implementation Steps

1. Add octicon paths to `icons.go` (copy from `primer/octicons` 16×16 viewBox).
2. Create `recordsCard` struct + `Filename()`.
3. Implement `SVG()`:
   - Compute 6 records via Phase-1 helpers (`time.Now()` for `accountAgeYears`).
   - Build `[]statRow` (or local equivalent) with formatted strings.
   - Empty-data branch: if `len(p.DailyContributionsAllTime) == 0`, label/value rows still render but values are `—`.
   - Emit header → 6 rows (same coords as stats card: `rowX=20, rowY0=55, rowDY=20, iconSize=12, valueX=320`) → footer.
4. Register in `allCards`.
5. `go build ./...` — confirm clean compile.
6. Run end-to-end render against the demo fixture if available, else inspect generated SVG manually.

## Success Criteria

- [ ] `go build ./...` clean
- [ ] Card filename `records.svg` registered and produced by `RenderAll`
- [ ] SVG validates (no XML parse errors) for all 65 themes
- [ ] Empty-data fixture produces a legible card (no panic, no missing values causing layout shift)

## Risk Assessment

- **Risk:** Row spacing collides at 6 rows × 20px DY + 55px Y0 = 175px; fits within 200px frame. **Mitigation:** verified against stats card (which has 7 rows at the same cadence). Safe.
- **Risk:** Long values (e.g. "612 in Mar 2026") overrun the value column at narrow themes. **Mitigation:** value text uses `text-anchor="end"` at `valueX=320` — overflows go left, never clip. If labels collide with values on a row, shorten label phrasing (e.g. "Best month" not "Best month ever").
- **Risk:** New octicon paths copied incorrectly (broken SVG). **Mitigation:** copy from `primer/octicons` source; smoke-test render in browser.
