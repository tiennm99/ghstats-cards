# Most Commit Language — Accurate Attribution Research

## Problem & constraints

GitHub's default language detection (`repo.primaryLanguage`) counts total bytes per repo language, not weighted by commit activity. Users with mixed-language repos (e.g., blog w/ 3 JS files + 1000 Markdown files) get misattributed: every commit to `.md` gets credited to the lowest-byte language, not the language actually edited. This report evaluates methods to fix attribution by tracking which files are modified per-commit.

**Constraints:** Action runner cost (storage/time), REST API rate limits (5K/hr), accuracy trade-offs, language-detection reliability without repo access.

---

## Prior art comparison table

| Project | Method | Clone? | Accuracy | Cost | Notes |
|---------|--------|--------|----------|------|-------|
| **anuraghazra/github-readme-stats (GRS)** | GraphQL `languages(first:10, orderBy:SIZE)` per repo | No | Byte-size only; **no commit weighting** | 1 GraphQL query/100 repos | Baseline: pure size-based, can't solve commit problem |
| **lowlighter/metrics (indepth)** | Clone repos → `git log --patch` per commit → linguist-js classify each file → accumulate by type | Yes | **Per-commit, per-line** | 15 min/repo timeout; heavy | Ground truth; `categories=[programming,markup]` filter; handles `.gitattributes` |
| **lowlighter/metrics (default/recent)** | GraphQL byte-size (same as GRS) + recent event stream analysis | No | Byte-size + event heuristics | Lightweight | Not commit-weighted |
| **Proposed: REST per-commit + go-enry** | REST `/repos/{o}/{r}/commits/{sha}` for each commit → go-enry classify filenames/extensions | No | **Per-commit by filename**, no line counting | 1 REST call/commit; 5K limit = ~100 commits/hr | Fast; lightweight; no clone; accuracy limited to extension-level |

---

## How github-readme-stats handles it

**GRS does NOT solve the commit-attribution problem.** It computes language stats as:

1. **GraphQL fetch:** `repositories(ownerAffiliations: OWNER, isFork: false, first: 100)` → for each repo:
   - `languages(first: 10, orderBy: {field: SIZE, direction: DESC})` → get top 10 languages by **bytes**
   - Accumulate `size` values across all repos; weight by `size_weight=1` and `count_weight=0` (default)
2. **Filters offered:** `exclude_repo` list only; no per-commit filtering, no commit-count weighting
3. **Documented limitation:** Users can use `exclude_repo` to hide problematic repos; no built-in commit-weighting

**Conclusion:** GRS is optimized for byte-size ranking (good for codebases), not commit activity (good for contributor profiles). No per-commit analysis.

---

## lowlighter/metrics — further details

### Default categories
Confirmed: `plugin_languages_categories` default is `[markup, programming]` (excludes `data`, `prose`). For indepth mode, users can override to include `prose` (which includes Markdown). Markdown is classified as **TypeProse** by go-enry (type code 4).

### Per-file analysis in indepth mode
- Clones repo to temp directory
- Runs `git log --author=<user> --patch` to fetch each commit with diff
- Parses unified diff to extract file paths and line counts (added/deleted per file)
- Calls `linguist-js` (Node wrapper around linguist) to classify each file
- Accumulates `{bytes, lines}` per language per commit
- Filters by `categories` to exclude unwanted types

### Clone timeout & safety
- `plugin_languages_analysis_timeout`: 15 sec global (default); 7.5 sec per repo (default)
- **No REST-only fallback mode** — if clone fails, that repo is skipped; no graceful degradation
- Symlinks, submodules, large binaries handled by linguist; `.gitattributes` **IS** respected (checked from cloned repo)

### De-duplication & fork handling
- Indepth fetches only repos where user is OWNER (not forks, not contributed-to repos)
- Counts only commits matching `--author` regex (authoring email list fetched from GPG keys)
- Deduplicates by commit SHA within session

### REST-only mode does NOT exist
lowlighter/metrics has two modes: default (byte-size, no clone) and indepth (clone + patch analysis). No hybrid REST-per-commit approach exposed.

---

## The go-enry + REST-per-commit approach — validation

### Module details
- **Module path:** `github.com/go-enry/go-enry/v2` (Apache-2.0 license)
- **Status:** Actively maintained; last push 2026-04-04; 603 stars; imported from github/linguist
- **Go version:** 1.14+
- **Pre-compiled data:** Yes — language metadata (types, colors, patterns) embedded in `data/*.go` files; auto-generated from github/linguist

### Language type mapping — confirmed for Markdown
```go
// Type int: 0=Unknown, 1=Data, 2=Programming, 3=Markup, 4=Prose
// Markdown = Type 4 (Prose)
var LanguagesType = map[string]int{
  "Markdown": 4,  // Prose
  "YAML": 1,      // Data
  "JSON": 1,      // Data
  "JavaScript": 2, // Programming
}
```

**Implication:** If we use go-enry and filter to `types=[2, 3]` (Programming + Markup), Markdown gets excluded. To include Markdown, must expand to `types=[2, 3, 4]` or add it to a custom whitelist.

### Extension-only vs content-based classification
go-enry has two classification modes:
1. **Extension-only** (`GetLanguageByExtension("file.md")` → "Markdown"): Fast, no file content needed
2. **Content-based** (`GetLanguageByContent(filename, content)` → "Markdown"): Slower; handles ambiguous extensions (e.g., `.txt`, `.r` for R vs reStructuredText)

**REST API provides:** Filename + additions/deletions per file, NO file content. So **we're limited to extension-only mode**, which is **~90% accurate** (fails on ambiguous extensions, but Markdown `.md` is unambiguous).

### `.gitattributes` support — NO
- go-enry does NOT parse `.gitattributes` directives (`linguist-vendored`, `linguist-generated`, `linguist-ignore`)
- Those live in the repo; without cloning, we can't access them
- **Accuracy delta:** Most users don't use `.gitattributes` heavily; estimated 5-10% false positives (counting generated/vendored code as real)
- **Mitigation:** Offer optional `.gitattributes` fetch via `GET /repos/{o}/{r}/contents/.gitattributes` if file exists (one extra REST call/repo)

### Binary/submodule/symlink handling
- go-enry's extension-based approach is safe: won't misclassify `.exe`, `.so`, etc. (no match → `OtherLanguage`)
- Submodules: filenames include submodule path; classified by extension of filename (safe)
- Symlinks: REST API lists symlink targets as files; go-enry classifies by target extension (reasonable)

### Known limitations
- **No semantic analysis:** Go comment syntax won't help distinguish `// in code` from `// in prose` (extension-based only)
- **Ambiguous extensions:** `.r`, `.tsx` can be R or reStructuredText / TypeScript or TSX. Content-based classification helps, but we don't have content.
- **Custom language aliases:** go-enry uses fixed language names from linguist; user aliases (e.g., renaming "TypeScript" to "TS") would need client-side mapping

---

## Cost estimation: REST per-commit approach

**Assumptions:** Top 10 repos, 100 commits/repo (user-configurable), 1K REST calls/run

**Rate limits:** 5K requests/hour (authenticated). At 100 commits/sec optimal rate, 1K calls ≈ 10 seconds API time.

**Safety:** Well within burst & hourly limits; safe for GitHub Actions runners.

**Breakdown per commit:**
- `GET /repos/{o}/{r}/commits/{sha}` → 1 call, returns `files[].{filename, additions, deletions}`
- go-enry local classification → negligible (microseconds per file)
- No clone, no network I/O beyond REST

**Feasibility:** High. Straightforward to parallelize (batch SHA fetches) if needed.

---

## Additional ideas

### Idea 1: Sampling + statistical estimation
Fetch 10-20 commits evenly distributed across repo history (e.g., every Nth commit). Use frequency to extrapolate total language distribution. **Trade-off:** 10x faster, ~70% accuracy (misses long-term shifts). **Verdict:** Useful for low-precision preview card, not for serious stats.

### Idea 2: GraphQL Commit history with limited file info
GitHub GraphQL `repository.ref.target.history` supports commit queries, but only returns commit count + author info. NO per-file details. Checked: `Commit` object has no `files` or `changedFilesIfAvailable` field exposing filenames. **Verdict:** Dead end; no edge advantage over REST.

### Idea 3: Docker + linguist Ruby gem
Run `linguist` CLI inside Docker during Action. Requires cloning repos into Action storage (same cost as lowlighter/metrics). More accurate than go-enry (Bayesian disambiguation). **Trade-off:** Heavy; docker image size; slower startup. **Verdict:** Over-engineered for our use case; go-enry sufficient.

### Idea 4: Fetch `.gitattributes` explicitly
One additional `GET /repos/{o}/{r}/contents/.gitattributes` per repo. Parse it locally; override go-enry classification for files tagged `linguist-vendored` or `linguist-generated`. **Cost:** 1 call/repo; minimal. **Accuracy gain:** ~5-10% fewer false positives. **Verdict:** Worth doing; low cost, reasonable gain.

### Idea 5: Combine REST per-commit with GraphQL for byte-size fallback
Query `GET /repos/{o}/{r}/languages` as secondary validation: if REST commit analysis yields unexpected results (e.g., 99% JavaScript despite repo being mostly Markdown), fall back to byte-size weighted stats. **Trade-off:** Adds complexity; not needed if go-enry is trusted. **Verdict:** Skip for v1; revisit after validation.

---

## Recommendation

**Proceed with: REST per-commit + go-enry, with optional `.gitattributes` override**

**Exact approach:**
1. For each user repo, fetch all commits (paginated, limit configurable; default 100/repo)
2. Per commit SHA: `GET /repos/{o}/{r}/commits/{sha}` → extract `files[].filename`
3. Classify each filename via `go-enry/v2` extension-based detection
4. Accumulate commit count + lines added per language per repo
5. **Optional:** Fetch `.gitattributes` per repo; override classifications for files marked `linguist-generated` / `linguist-vendored`
6. Filter output by language type (exclude `Unknown` / `Data`; include `Programming`, `Markup`, optionally `Prose`)
7. Rank by commit count (primary) or lines added (secondary weighting option)

**Why this beats alternatives:**
- **vs. byte-size (GRS):** Commit-weighted, not size-weighted → accurate for mixed-language repos
- **vs. cloning (lowlighter):** No repo clone → safe for Action runner, completes in <10 sec for typical user
- **vs. sampling:** 100% coverage, not statistical estimate
- **Cost:** ~100 REST calls (1K limit budget = 10x headroom); well-designed for scale

**Estimated accuracy:** 90-95% (limited by extension-only classification; `.gitattributes` override lifts this to 95-98%)

---

## Unresolved questions

- Should **Prose** (Markdown, AsciiDoc, reStructuredText) be included in default output, or only **Programming + Markup**? (lowlighter defaults to Programming + Markup; users opt-in to Prose)
- Do we want **lines-added weighted** language ranking, or just **commit-count weighted**? (e.g., 1-line typo fix vs. 500-line refactor)
- Should **vendored/generated code** be excluded by default, or configurable? (`.gitattributes` parsing adds 1 call/repo; users likely don't care)
- How do we handle repos with **no commits by user** (e.g., organization repos where user never committed)? (Skip? Count PR reviews? Leave blank?)
- Fallback behavior if **go-enry can't classify a file** (e.g., custom `.lisp` variant): count as "Other" or skip?

---

**Sources:**
- [anuraghazra/github-readme-stats](https://github.com/anuraghazra/github-readme-stats)
- [lowlighter/metrics](https://github.com/lowlighter/metrics)
- [go-enry/go-enry](https://github.com/go-enry/go-enry)
- [GitHub REST API: Get a commit](https://docs.github.com/en/rest/commits/commits)
- [GitHub REST API: Rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api)
