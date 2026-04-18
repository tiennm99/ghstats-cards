# System Architecture

## Runtime shape

One process, three phases: **flag parsing → data fetch → SVG render**.

```
┌───────────┐   ┌──────────────────┐   ┌─────────────────┐   ┌──────────────┐
│ flag / env│──►│  internal/github │──►│  internal/card  │──►│  output/*.svg│
│  parsing  │   │  (GraphQL only)  │   │  (pure render)  │   │   per theme  │
└───────────┘   └──────────────────┘   └─────────────────┘   └──────────────┘
                         ▲                       ▲
                         │                       │
                    api.github.com        internal/theme
```

No database, no cache, no background workers. Stateless CLI; Action runtime just sets environment variables + runs the binary.

A root `context.Context` is built in `main.go` with an overall deadline (`-timeout`, default 30m) and cancelled on `SIGINT`/`SIGTERM`. Every fetcher and HTTP request inherits it so a slow run aborts cleanly instead of draining the 6h Action budget.

## Data-fetch sequence

```
main.go
  │
  ▼
FetchProfile(ctx, login, opts)
  │  profileQuery × N pages (owned repos, STARGAZERS desc, 100/page)
  │  yields: Profile.{identity, stars, forks, PRs, issues,
  │                   TopRepos, ReposByLanguage,
  │                   ContributionYears,
  │                   DailyContributions (last year),
  │                   TotalCommits (last year)}
  │
  ▼
FetchContributionsAllTime(ctx, profile, opts)
  │  contributionYearQuery × len(ContributionYears)
  │  per year: totalCommitContributions +
  │            contributionCalendar.weeks +
  │            commitContributionsByRepository(maxRepositories: 100)
  │  yields: SeedRepos (deduped),
  │          DailyContributionsAllTime,
  │          TotalCommitsAllTime
  │
  ▼
FetchProductive(ctx, profile, profile.SeedRepos, loc, commitsPerRepo)
  │  commitHistoryQuery × (#seeds × pages)
  │  per commit: t = committedDate in loc
  │              ProductiveAllTime[t.Hour]++  + language votes
  │              if t.After(yearAgo): Productive[t.Hour]++ + language votes
  │  yields: Productive, ProductiveAllTime,
  │          CommitsByLanguage, CommitsByLanguageAllTime
  │
  ▼
card.RenderAll(profile, theme, outDir)  ×  len(themes)
```

## GraphQL queries

All three queries live in `internal/github/queries.go`.

| Query | Purpose | Cost estimate |
| --- | --- | --- |
| `profileQuery` | Profile identity + totals + owned repos + last-year calendar | 1–10 calls (100 repos/page × ≤10 pages safety cap) |
| `contributionYearQuery` | Per-year calendar + seed list | 1 call per active year (typically 1–10) |
| `commitHistoryQuery` | Authored commits on default branch | 1 call per 100 commits per seed repo |

Typical run (8 active years, 30 seed repos, avg 50 commits each):
- profile: 1 call
- year loop: 8 calls
- commit history: 30 × 1 = 30 calls
- **≈ 39 GraphQL calls, 0 REST calls**

## Attribution model

Language attribution for the "most commit language" card is **byte-weighted**:

```
for each repo R:
    total_bytes = Σ R.languages[*].bytes   // precomputed once per repo
    for each commit C in R:
        for each (lang, bytes) in R.languages:
            commits_by_lang[lang] += scaleFactor × bytes / total_bytes
```

Implementation in `internal/github/productive.go:attributeCommit`. The per-repo byte total is hoisted out of the commit loop so the hot path doesn't re-sum language edges for every commit. `scaleFactor = 10_000` preserves fractional precision in int64 storage — percentages rendered in the card are unaffected by magnitude.

Known distortion: linguist excludes prose types (Markdown, AsciiDoc, reST) from byte counts. Blog-style repos with 95% Markdown and 5% JS still attribute all commits to JS. Future fix: per-commit REST file classification via `-accurate-languages` (see roadmap).

## SVG generation

Each card produces a self-contained SVG with:
- Card frame (rounded rect, theme background, theme stroke + opacity)
- Title (top-left, theme title color)
- Content layer (chart elements, text, legend)

Shared primitives:
- `renderDonutCard(title, stats, theme)` — pie slices via polar arc math + legend with color swatches. Single-slice case (one language at 100%) renders as two concentric `<circle>` elements instead of an arc, since SVG's `A` command from point P back to P draws nothing.
- `renderProductiveTime(title, hours, theme)` — 24 bars + both axes + tick math from `niceTicks`
- `renderContributions(title, days, theme)` — monthly aggregation, Catmull-Rom → cubic Bezier area path, two-sided Y axis

Catmull-Rom control-point math: for each segment `P_i → P_{i+1}`,
```
C1 = P_i + (P_{i+1} - P_{i-1}) / 6
C2 = P_{i+1} - (P_{i+2} - P_i) / 6
```
Tension = 0.5 (d3's default).

## Theme model

`theme.Theme` is a pure-data struct — no methods. Cards pull `t.Background`, `t.Text`, `t.Title`, `t.Accent`, `t.Muted`, `t.Stroke`, `t.StrokeOpacity`. The 61 palettes live in a map keyed by snake_case ID.

Light themes (`default`, `github`, `nord_bright`, etc.) use `StrokeOpacity: 1` with a visible stroke color; dark themes often use `StrokeOpacity: 0` or a stroke that blends into the background.

## Failure modes

| Fault | Behavior |
| --- | --- |
| Empty `-user` | Exit 2, usage printed |
| Unknown theme | Exit 2, suggests `-list-themes` |
| GraphQL 4xx/5xx | Error wrapped with HTTP status and truncated (UTF-8-safe) body |
| Primary rate limit (429 / 403 + remaining=0) | Sleep up to 5 min honoring `Retry-After` / `X-RateLimit-Reset`, retry once; longer windows surface as error |
| Per-year query returns nil user | Warn to stderr; other years still contribute |
| `FetchProductive` network error | Warn to stderr; partial data rendered |
| Unknown timezone | Warn to stderr; fall back to UTC |
| Overall timeout (`-timeout`) or Ctrl-C | `ctx` cancels in-flight requests; partial data may render |
| User with 0 commits | Card renders "No data available" |

## Extension points

- **New card**: implement `Card` interface, add to `allCards` in `card.go`.
- **New theme**: add entry to `themes` map in `theme.go`.
- **New fetcher mode** (e.g., REST per-commit): add a new method on `*Client`, call from `main.go`, wire to new `Profile` fields.
