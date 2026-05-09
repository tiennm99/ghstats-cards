---
phase: 1
title: "Records computation helpers"
status: completed
priority: P1
effort: "1h"
dependencies: []
---

# Phase 1: Records computation helpers

## Overview

Add pure functions that derive the 6 records from `Profile` data already in memory. Live in `internal/card/records.go` next to the card itself — no new public package surface needed.

## Requirements

**Functional**
- `peakDay(days []DailyContribution) (count int, date time.Time)` — argmax; ties → earliest date
- `peakMonth(days []DailyContribution) (count int, ym time.Time)` — sum per `YYYY-MM`, argmax; ties → earliest month
- `firstActiveDay(days []DailyContribution) time.Time` — first day with `Count > 0`; zero time if none
- `activeDaysCount(days []DailyContribution) int` — count of days where `Count > 0`
- `accountAgeYears(createdAt, now time.Time) float64` — fractional years, 1 decimal
- `languagesUsed(stats []LangStat) int` — `len(stats)` (already deduped upstream)

**Non-functional**
- O(n) over `DailyContributionsAllTime` for all daily-derived records (single pass acceptable)
- Zero allocations for argmax helpers (just iterate)
- All helpers package-private (`unexported`) — only the card uses them

## Architecture

Single file `internal/card/records.go` exporting the `recordsCard` struct (Phase 2) and these helpers. No mutation of `Profile`. Empty-input handling: return zero values; the card layer decides what to display.

## Related Code Files

- Create: `internal/card/records.go` (helpers section)
- Read for context: `internal/github/model.go` (Profile fields), `internal/card/stats.go` (card pattern)

## Implementation Steps

1. Create `internal/card/records.go` with package-private helpers.
2. Implement `peakDay`: iterate, track `(maxCount, earliestDate)`; tie-break on earlier date.
3. Implement `peakMonth`: bucket sum into a `map[time.Time]int` keyed by `time.Date(year,month,1,...)`; argmax with same tie-break.
4. Implement `firstActiveDay`: first index where `Count > 0`; assumes input is chronological (matches existing fetcher).
5. Implement `activeDaysCount`: counter of `Count > 0`.
6. Implement `accountAgeYears`: `now.Sub(createdAt).Hours() / 24 / 365.25`, round to 1 decimal.
7. Implement `languagesUsed`: trivial wrapper for symmetry/testability.

## Success Criteria

- [ ] All 6 helpers compile (`go build ./...`)
- [ ] Helpers produce correct results on hand-rolled fixtures
- [ ] Empty-data inputs return zero values without panic

## Risk Assessment

- **Risk:** `peakMonth` map ordering non-deterministic — Go map iteration is randomized, so two months tied at the same count may produce different "winners" across runs. **Mitigation:** scan keys, sort by date ascending, then argmax on first occurrence. Alternative: track `(count, ym)` during the bucket-fill pass instead of a second pass.
- **Risk:** `accountAgeYears` rounding inconsistent with existing profile-card "age" string. **Mitigation:** match `internal/card/profile.go` convention (check there before implementing).
