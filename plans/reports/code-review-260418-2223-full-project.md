# Full-Project Code Review

Scope: all Go source, `action.yml`, `entrypoint.sh`, `Dockerfile`, workflows. ~2300 LOC. Adversarial pass included.

Verification harness used: Go compile + tests pass clean (`go vet ./... && go test ./...`). Real-data render against `tiennm99` with token — all 9 cards produced valid SVGs. One bug reproduced with a standalone probe.

## Verdict

Code quality is high for the scope. One real rendering bug, several correctness and hygiene issues, and a meaningful test-coverage gap. Nothing blocks merge; prioritize the **Important** list before next tag.

## Severity counts

| Critical | Important | Nice-to-have |
|---:|---:|---:|
| 0 | 6 | 10 |

---

## Important

### I1 — Donut chart renders empty when there is only 1 slice ⚠️
**File**: `internal/card/donut_chart.go:60-79`

For a user with a single language at 100%, `angle = 2π` → `start == end` → SVG `A` command degenerates to zero-length arc. The donut shows nothing.

Reproduced with a standalone probe:
```
start=(380, 50)
end=(380, 50)      # identical → empty arc
same=true
```

**Fix**: when there's exactly one slice (or when `angle >= 2π - ε`), render a full ring via two half-arcs, a `<circle>` with a stroke, or a punched-out clip path. Smallest change:
```go
if len(stats) == 1 {
    // Full ring: outer circle fill + inner background circle
    fmt.Fprintf(&b, `<circle cx=... fill="%s"/><circle cx=... fill="%s"/>`, slice.Color, t.Background)
    ...
}
```

### I2 — `FetchContributionsAllTime` silently drops failed years
**File**: `internal/github/contributions_all_time.go:62-64`

```go
if resp.User == nil {
    continue
}
```
If GraphQL returns an error JSON per year (e.g. permission issue, transient 5xx that somehow bypassed the HTTP error check), the year is skipped with zero logging. All 8 years of a user's history could silently vanish — cards render "No data available" with no diagnostic.

**Fix**: log at warn level when `resp.User == nil` after a successful response; or return a wrapped error so `main.go` can surface a partial-data warning.

### I3 — Stale comment on `FetchOptions` contradicts live behavior
**File**: `internal/github/profile.go:61-62`

```go
// FetchOptions tunes which repos contribute to the profile's aggregates.
// All defaults are conservative (no forks, no private) so public-facing
// READMEs don't accidentally leak work-repo signal.
```
Defaults are now **true/true** (commit `514195c`). Zero-value `FetchOptions{}` still gives `false/false`, so API callers using `FetchOptions{}` get one behavior while CLI callers get another. Subtle surprise.

**Fix**: either update the comment to reflect new posture, or flip the zero-value semantics (e.g., `ExcludeForks bool` / `ExcludePrivate bool`) so the package-level default matches the CLI. The former is simpler.

### I4 — `TestRenderAll` no longer verifies XML escape
**File**: `internal/card/card_test.go:17,63`

The test sets `Bio: "Test & <bio>"` and asserts the raw string doesn't appear in any rendered SVG. But `Bio` is **not rendered** anywhere after the profile-card redesign — it was removed along with the github-row. The assertion is trivially true for every future change.

Additionally, `TestRenderAll` only checks each file starts with `<svg` — it does not validate content. A card that renders an empty shell would pass.

**Fix**: set `Name: "Alice & <bob>"` (Name **is** rendered, in the title) and keep the assertion. Add a golden-file comparison per card for at least one theme to catch silent regressions.

### I5 — Release workflow doesn't gate on tests
**File**: `.github/workflows/release.yml`

Tagging `v1.2.3` runs the release pipeline immediately — no `go test ./...`, no `needs: [ci]`. A broken `main` tagged accidentally ships broken binaries + Docker image.

**Fix**: add a pre-release test job:
```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v6
    - uses: actions/setup-go@v6
      with:
        go-version: "1.26"
    - run: go vet ./... && go test ./...
docker:
  needs: [test]
  ...
binaries:
  needs: [test]
  ...
```

### I6 — `attributeCommit` recomputes `total` per-commit
**File**: `internal/github/productive.go:111-115` + `:86-92` (caller)

```go
func attributeCommit(repo RepoInfo, ...) {
    var total int64
    for _, l := range repo.Languages {
        total += l.Bytes
    }
    ...
}
```
Called once per commit. For 500 commits × 6-20 languages per repo, that's 3000–10000 redundant additions per repo. Not a runtime problem, but it's a correctness smell: if `repo.Languages` were ever mutated between calls, you'd get inconsistent totals.

**Fix**: precompute `total` once per repo, pass it in or cache on `RepoInfo`. Clean separation of loop-invariant work.

---

## Nice-to-have

### N1 — `joinErrs` reinvents `strings.Join`
**File**: `internal/github/client.go:102-110`

Custom loop concatenates with `"; "`. Identical to `strings.Join(ss, "; ")`. Replace.

### N2 — No total timeout / `context.Context`
**File**: `internal/github/client.go:26`, all fetchers

`http.Client.Timeout = 30s` applies per request. A fetch with 50 pages × 30s worst-case = 25 minutes. No way to set an overall budget or cancel on signal. Add `ctx context.Context` to `Client.query` and fetcher methods; respect `<-ctx.Done()` in pagination loops.

### N3 — `truncate` may split UTF-8 mid-sequence
**File**: `internal/github/client.go:95-100`

`string(b[:n])` at an arbitrary byte boundary can leave a half-character. Purely cosmetic in error messages; fine to ignore.

### N4 — Docker base images pinned to major, not digest
**File**: `Dockerfile:1,8`

`golang:1.26-alpine` and `alpine:3.21` are mutable tags. A compromised Alpine push affects builds. Pin to `@sha256:…` for hermetic builds. Acceptable risk for an OSS tool.

### N5 — Third-party GHA actions pinned to major
**File**: `.github/workflows/release.yml`

`docker/build-push-action@v6`, `softprops/action-gh-release@v2`, `actions/checkout@v6` — mutable. Pin to SHA for supply-chain safety. Same caveat as N4.

### N6 — No rate-limit header inspection
**File**: `internal/github/client.go:63-73`

`x-ratelimit-remaining` / `x-ratelimit-reset` are ignored. When near zero, the next call returns 403 and we error out. Cheap improvement: parse headers and sleep until reset.

### N7 — `xAxisLabelVisible` rules can drop the penultimate label without the cosmetic expectation
**File**: `internal/card/contributions.go:158-160`

```go
if n-1-i < stride/2 {
    return false
}
```
For some `(n, stride)` pairs this drops a label that would have been `stride` apart. Result: 5 labels where you expected 6. Minor visual quirk; not a bug.

### N8 — `Profile.TotalContributions` is misnamed
**File**: `internal/github/model.go` (via `profile.go:118`)

Set from `ContributionCalendar.TotalContributions + RestrictedContributionsCount`, which is a **one-year** total, not lifetime. Field name suggests lifetime. Rename to `TotalContributionsLastYear` or compute from the AllTime loop.

### N9 — Stats card's "Contributed to (non-fork)" stays capped at the top-N
**File**: `internal/card/stats.go` + `profile.go:113`

`TotalContributedTo = u.RepositoriesContributedTo.TotalCount` queries with `first: 1`, so it only returns the **count** — that's fine. But the label says "(non-fork)" while the query doesn't actually filter by fork. Drop the "(non-fork)" qualifier or add `contributionTypes` filter that excludes forks.

### N10 — `output/` in .gitignore exception is narrow but fragile
**File**: `.gitignore`

```
output/*
!output/dracula/
```
If someone runs with `-out=output/` and a theme named `dracula`, their local changes overlay the committed sample. Usually fine. If someone runs `-themes all`, `output/*` blocks every theme dir except dracula — working as intended. No action needed; noted for future refactors.

---

## Adversarial pass — things that don't break

Tried and confirmed safe:

| Attack | Result |
| --- | --- |
| XML injection via `Bio`, `Name`, `Company`, `Location`, `Website`, language names | All flow through `escapeXML` — safe |
| Theme ID path traversal (`-themes ../../etc`) | `theme.Lookup` rejects unknown IDs before `filepath.Join` — safe |
| GraphQL injection via `$login`, `$owner`, `$repo` | Variables passed separately from query string — safe |
| Shell injection in `entrypoint.sh` via user inputs | All variables double-quoted; no `eval` — safe |
| Token exposure in logs | `entrypoint.sh:31` echoes user/themes/out only; no token echo — safe |
| Token exposure in error messages | HTTP errors truncate body to 500 bytes; GitHub bodies don't echo tokens — safe |
| Integer overflow in `scaleFactor * bytes / total` | Worst realistic case ~10^14, well under int64 max — safe |
| Panic from `Productive[tl.Hour()]` | `Hour()` always returns 0-23 — safe |
| NaN from `angle := 2 * math.Pi * value / total` when total=0 | len(stats)>0 check guarantees one non-zero value, so total>0 in practice. But... if all values happen to be 0, we'd produce NaN. Unlikely (sortLangStats wouldn't generate 0-valued entries), but defensive check wouldn't hurt. |
| Resource exhaustion via huge user | `maxPages = 10` in FetchProfile (1000 repos cap); `maxRepositories: 100` per year in seed query. Bounded. |
| Race conditions | Single-goroutine — no races possible |
| Division by zero in `xAxisLabelVisible` stride calc | Early return for `n <= xLabelTarget` prevents stride=0 — safe |

---

## Testing gaps (summary)

| Missing test | Motivation |
| --- | --- |
| Single-slice donut | Catches bug I1 |
| `catmullRomLinePath([1 point])` / `[]` | Verifies early returns |
| Half-hour timezone (`Asia/Kolkata`) in `utcOffsetLabel` | Verifies `%+.2f` formatting |
| `sortLangStats` with ties and empty color map | Already partially tested |
| Empty `DailyContributions` → "No data available" path | Already partially tested indirectly |
| Golden SVG comparison per card for one theme | Catches regressions across refactors |

---

## Not issues

- No external Go deps. Good.
- `filepath.Join(outDir, t.ID)` — `t.ID` sourced from the themes map keys, validated through `theme.Lookup`. Safe.
- GraphQL query strings are compile-time constants. No injection surface.
- Commit-time error paths (`git add`, `git commit`, `git push`) run in Action workspace only and fail closed.

---

## Unresolved questions

- Should we land `-accurate-languages` (REST-per-commit + enry) as the primary fix for Markdown-blog attribution, or invest in `-deep` (partial bare clone) first? Roadmap has both.
- Is the single-slice donut case (I1) rare enough in practice to accept a compact fix, or worth refactoring to render all donuts with a full-ring base + pie slices on top (more robust, slightly more code)?
- Release workflow: run tests in matrix (cross-platform) or on linux only before shipping? Linux-only is practical; matrix catches platform bugs but triples minutes.
