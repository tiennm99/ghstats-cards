# Codebase Summary

## Layout

```
ghstats/
├── main.go                              # CLI entry point; wires flags → fetchers → renderers
├── action.yml                           # GitHub Action metadata
├── entrypoint.sh                        # Action runtime; maps INPUT_* env → CLI flags
├── Dockerfile                           # Multi-stage build for the Action image
├── go.mod                               # Module declaration; no external deps
├── internal/
│   ├── github/                          # GraphQL client + fetchers + models
│   │   ├── client.go                    # HTTP POST to /graphql, error decoding
│   │   ├── queries.go                   # profileQuery, commitHistoryQuery, contributionYearQuery
│   │   ├── model.go                     # Profile, RepoInfo, LangStat, LangEdge, DailyContribution
│   │   ├── profile.go                   # FetchProfile — user + owned repos + stats + calendar
│   │   ├── productive.go                # FetchProductive — commit history → hour histogram + lang buckets
│   │   ├── contributions_all_time.go    # FetchContributionsAllTime — per-year loop → seed list + daily series
│   │   └── profile_test.go              # sortLangStats tiebreak
│   ├── card/                            # SVG renderers; one file per card
│   │   ├── card.go                      # Card interface, RenderAll, allCards slice
│   │   ├── svg.go                       # escapeXML, formatInt, header, footer
│   │   ├── axis.go                      # niceTicks (d3-style 1/2/5 × 10^k), formatTick
│   │   ├── icons.go                     # Octicon path strings
│   │   ├── profile.go                   # profile-details
│   │   ├── repos_per_language.go        # repos-per-language
│   │   ├── most_commit_language.go      # most-commit-language
│   │   ├── most_commit_language_all_time.go  # most-commit-language-all-time
│   │   ├── stats.go                     # stats
│   │   ├── productive.go                # productive-time (+ all-time)
│   │   ├── contributions.go             # contributions (+ all-time)
│   │   ├── donut_chart.go               # renderDonutCard — shared by language cards
│   │   └── card_test.go                 # Rendering + escape + format tests
│   └── theme/
│       └── theme.go                     # 65-palette map ported from github-profile-summary-cards
├── .github/workflows/
│   ├── ci.yml                           # go vet + go test on push/PR
│   ├── release.yml                      # GHCR image + cross-platform binaries on tag
│   └── demo.yml                         # Renders every theme for the repo owner on push to main
├── docs/                                # This directory
├── plans/                               # Research reports + implementation plans
├── demo/                                # Auto-generated gallery — every card × every theme + README
└── output/dracula/                      # Sample committed; other themes gitignored
```

## Module responsibilities

### `internal/github`

All network I/O. Exposes a `*Client` with three fetchers; every call takes a `context.Context` so pagination aborts cleanly on timeout or Ctrl-C:

| Fetcher | Input | Populates |
| --- | --- | --- |
| `FetchProfile(ctx, login, opts)` | username, visibility flags | Profile basics, totals, owned-repos aggregation, last-year daily calendar, `TopRepos` |
| `FetchContributionsAllTime(ctx, p, opts)` | Profile | `SeedRepos`, `DailyContributionsAllTime`, `TotalCommitsAllTime` |
| `FetchProductive(ctx, p, repos, loc, cap)` | Profile + seed + tz + cap | `Productive`, `CommitsByLanguage`, `ProductiveAllTime`, `CommitsByLanguageAllTime` |

Call order in `main.go`: Profile → AllTime → Productive. `Client.query` handles GitHub rate limits transparently — on 429 or 403 with `X-RateLimit-Remaining: 0`, it honors `Retry-After` / `X-RateLimit-Reset` (capped at 5 minutes) and retries once.

### `internal/card`

Pure rendering. Every card implements the `Card` interface:

```go
type Card interface {
    Filename() string
    SVG(*github.Profile, theme.Theme) ([]byte, error)
}
```

`RenderAll` iterates `allCards`, writes each to `<outDir>/<themeID>/<Filename>`.

Shared helpers:
- `renderDonutCard` — language donut + legend (used by 3 language cards)
- `renderProductiveTime` — 24h bar chart (used by both productive cards)
- `renderContributions` — smooth area chart (used by both contributions cards)
- `header`, `footer` — SVG chrome
- `niceTicks`, `formatTick` — axis math

### `internal/theme`

Static map of 65 themes. Each theme specifies title/text/background/stroke/accent/muted plus `StrokeOpacity` for correct light-theme borders.

## Card ↔ data flow

```
profileQuery ─────► Profile.{identity, owned repos, totals, last-year calendar}
                                │
contributionYearQuery ─┬──► SeedRepos + DailyContributionsAllTime + TotalCommitsAllTime
                       │
                       └─ seed into ─►
                                │
commitHistoryQuery ──► Productive + CommitsByLanguage (+ AllTime variants)
                                │
                                ▼
                          9 SVG files per theme
```

## Test coverage

- `internal/card/card_test.go` — `RenderAll` produces 9 valid SVGs; XML escape through real render pipeline; `formatInt` cases; `TestDonutSingleSlice` (guards the empty-arc regression); `TestDonutEmpty` (no-data fallback).
- `internal/github/profile_test.go` — `sortLangStats` ordering and tiebreak.
- `main_test.go` — `TestUTCOffsetLabel` covers UTC, Asia/Saigon, half-hour (Kolkata), quarter-hour (Kathmandu) zones.

No network-touching tests; real runs verified via `-token` + local build.

## Naming conventions

- Go files use snake_case for multi-word names (`repos_per_language.go`, `contributions_all_time.go`).
- Cards' `Filename()` returns a plain kebab-case name (`profile-details.svg`, `most-commit-language-all-time.svg`, …). Embedders reference by name, so no numeric prefix is needed.
- Themes in snake_case to match upstream (`github_dark`, `nord_bright`).
