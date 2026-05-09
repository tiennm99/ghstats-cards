---
title: Add records card
status: completed
created: 2026-05-09
slug: records-card
---

# Add records card

A new SVG card surfacing **personal extremes & milestones** rather than aggregates. Mirrors `stats.svg` layout (icon + label + right-aligned accent value).

## Goal

Show 6 records that tell a story arc — **origin → tenure → peaks → breadth**:

| Row | Record | Source |
|---|---|---|
| 1 | Best day (count + date) | argmax over `DailyContributionsAllTime` |
| 2 | Best month (count + `YYYY-MM`) | bucket-sum over `DailyContributionsAllTime`, argmax |
| 3 | First contribution (date) | first non-zero day in `DailyContributionsAllTime` |
| 4 | Active days (lifetime) | count non-zero days in `DailyContributionsAllTime` |
| 5 | On GitHub (years) | `now − Profile.CreatedAt` |
| 6 | Languages used | `len(Profile.CommitsByLanguageAllTime)` |

**No new GraphQL queries.** All data already populated by existing fetchers.

## Non-goals

- Card #2 variant for last-year records (low-value duplication)
- Streak/top-repo/peak-year (already covered by other cards)
- Configurable record set (YAGNI)

## Phases

| # | Phase | Status | File |
|---|---|---|---|
| 1 | Records computation helpers | completed | [phase-01-records-computation.md](./phase-01-records-computation.md) |
| 2 | SVG card rendering | completed | [phase-02-svg-card-rendering.md](./phase-02-svg-card-rendering.md) |
| 3 | Tests + README | completed | [phase-03-tests-and-readme.md](./phase-03-tests-and-readme.md) |

## Key dependencies

- Existing fetched data: `Profile.DailyContributionsAllTime`, `Profile.CommitsByLanguageAllTime`, `Profile.CreatedAt`
- Existing card pattern: `internal/card/stats.go` (icon + label + value rows)
- Existing icon set: `internal/card/icons.go` (may need 1-3 new octicon paths)

## Definition of done

- `records.svg` renders for all 65 themes via demo workflow
- Card registered in `allCards` (renders alongside other cards)
- Unit tests cover record-extraction edge cases (empty data, single-day, ties)
- README updated with new row in card table + dracula preview
