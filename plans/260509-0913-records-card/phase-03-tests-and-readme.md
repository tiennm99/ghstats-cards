---
phase: 3
title: "Tests + README"
status: completed
priority: P2
effort: "45m"
dependencies: [1, 2]
---

# Phase 3: Tests + README

## Overview

Cover the 6 record helpers with unit tests, add a card-level smoke test, and document the new card in `README.md` so it shows up in the gallery preview.

## Requirements

**Functional**
- Helpers tested for: empty input, single-day, ties, all-zero counts
- Card-level test asserts `records.svg` filename + key strings appear (label rows, formatted values)
- README updated:
  - New row in card table at the same row index where it lands in `allCards`
  - Dracula preview cell pointing at `./demo/dracula/records.svg`

**Non-functional**
- Test file follows existing convention (`*_test.go` in `internal/card`)
- No flaky time-based tests — pass `now` explicitly into `accountAgeYears`

## Architecture

```
internal/card/records_test.go
├── TestPeakDay (incl. ties → earliest)
├── TestPeakMonth (incl. ties → earliest, multi-month)
├── TestFirstActiveDay (incl. all-zero → zero time)
├── TestActiveDaysCount
├── TestAccountAgeYears (fixed createdAt vs fixed now)
├── TestLanguagesUsed
└── TestRecordsCardSVG (smoke: contains title + 6 row labels)
```

## Related Code Files

- Create: `internal/card/records_test.go`
- Modify: `README.md` (card table + dracula gallery cell)
- Read for context: `internal/card/weekday_start_test.go` (existing test patterns)

## Implementation Steps

1. Write helper unit tests with hand-rolled `[]DailyContribution` fixtures.
2. Write card-level test asserting:
   - SVG bytes contain `Records (all time)` title
   - SVG contains 6 row labels (`Best day`, `Best month`, ...)
   - Empty-fixture variant still renders without panic
3. `go test ./...` — confirm pass.
4. Update README:
   - Card table: add row `15` (or whatever index) with description
   - Gallery: add `<td><img src="./demo/dracula/records.svg" alt="records" /></td>` to dracula preview block
5. Trigger demo workflow regen (or rely on next push).

## Success Criteria

- [ ] All new tests pass (`go test ./...`)
- [ ] README renders correctly (card table + gallery image link)
- [ ] Demo workflow produces `records.svg` for all 65 themes after merge

## Risk Assessment

- **Risk:** Card-level snapshot test gets brittle as themes evolve. **Mitigation:** assert on text fragments only, never on full byte equality.
- **Risk:** README gallery layout breaks if the new image cell forces an odd column count. **Mitigation:** keep dracula gallery's `<table>` row count even (pair with another card or span 2 cols).
