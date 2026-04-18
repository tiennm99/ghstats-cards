# Project Roadmap

## Phase 0 — Skeleton (✅ done)

- Module layout, flag parsing, placeholder SVG renderers.

## Phase 1 — Five core cards (✅ done)

- Profile details, repos-per-language, most-commit-language, stats, productive-time.
- GraphQL profile query + per-repo commit history.
- Docker-based Action wrapper, release workflow, 61-theme palette.

## Phase 2 — Chart quality (✅ done)

- Match github-profile-summary-cards visual style (donuts, 24h bar chart, proper axes).
- Octicon labels on profile + stats cards.
- Smooth area chart for contributions (Catmull-Rom → cubic Bezier).

## Phase 3 — All-time variants (✅ done)

- Unified commit-history fetch splits into last-year and all-time buckets.
- Per-year `contributionsCollection` loop yields `DailyContributionsAllTime` + `TotalCommitsAllTime`.
- Three new cards: 6-most-commit-language-all-time, 7-productive-time-all-time, 8-contributions-all-time.
- Stats card gains a lifetime commits row.

## Phase 4 — Accurate repo sampling (✅ done)

- Seed list built from `commitContributionsByRepository` across every active year.
- `-include-forks` / `-include-private` visibility flags (default off).
- `-top-repos` demoted to an optional cap (default 0 = unlimited).
- Commit-history query takes `$owner` so forks and non-owned repos are probeable.

## Phase 5 — Byte-weighted attribution (✅ done)

- Each commit distributes fractionally across repo's language bytes, not just primary.
- Improves mixed-code repo accuracy; still inaccurate for Markdown-heavy repos (linguist prose-exclusion).

---

## Phase 6 — Per-commit file classification (planned)

**Goal**: fix the Markdown-blog misattribution case (and any repo where linguist's byte view disagrees with what files user actually edited).

**Approach**: `GET /repos/{owner}/{repo}/commits/{sha}` per commit → classify each file with `go-enry`. Weight by `additions + deletions`.

**Cost**: ~1 REST call per commit. At current defaults (30 seed repos × 500 commits = 15,000 commits worst case) this is heavy — needs `-accurate-languages` opt-in flag, schedule weekly not daily.

**Research**: see `plans/reports/researcher-260418-2001-accurate-language-stats.md`.

**Status**: designed, not implemented.

## Phase 7 — Partial bare clone for lifetime all-repo stats (planned)

**Goal**: lifetime language stats across **every** repo a user has committed in, without the 500-commits-per-repo cap.

**Approach**: `git clone --filter=blob:none --bare` per seed repo + `git log --author --numstat` → go-enry.

**Cost**: ~5% of full-clone disk (trees only, no blobs); 3–5 minutes runtime for 100 repos; zero REST calls.

**Trade-off**: needs disk + git binary on runner. Lowlighter/metrics' indepth mode does similar but clones full blobs; we'd skip those.

**Status**: researched only; behind `-deep` flag when landed.

## Phase 8 — User-configurable repo exclusion (planned)

**Goal**: let users drop throwaway repos (experiments, forks they stashed) from stats without disabling forks globally.

**Approach**: `-exclude-repo owner1/name1,owner2/name2` flag. Filter seed list before probing.

**Cost**: negligible (client-side filter).

**Status**: pending user demand.

## Phase 9 — Expand ownerAffiliations (planned)

**Goal**: catch work done in org repos where user is a collaborator, not owner (e.g., company monorepos).

**Approach**: expose `-affiliations OWNER,COLLABORATOR,ORGANIZATION_MEMBER` flag. Requires thinking about whether to *display* private org work on a public profile card.

**Status**: blocked on deciding the privacy default.

---

## Known limitations (not roadmap items — by design)

| Limitation | Reason |
| --- | --- |
| Markdown/prose excluded from byte counts | Linguist's default; we defer to linguist |
| No real-time API | Scope: scheduled batch renderer, not a server |
| No WakaTime integration | Out of scope — WakaTime cards already exist (athul/waka-readme, anmol098/waka-readme-stats) |
| No heatmap (7×24) variant of productive time | Simplified to 24-hour bar chart to match reference project |
| Hard width of 500 px per card | Keeps README layout predictable; customizing width would cascade through every chart math |

## Tracked research reports

All in `plans/reports/`:
- `researcher-260418-2001-accurate-language-stats.md` — metrics vs GRS vs go-enry feasibility
- `researcher-260418-2012-profile-stats-survey.md` — follow-up survey across 6 more tools
- `analysis-260418-2140-most-commit-language-all-time.md` — hand-reconstruction of tiennm99's card output, showing exactly why each language lands where
