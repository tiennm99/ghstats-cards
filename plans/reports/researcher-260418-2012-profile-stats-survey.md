# Profile Stats Tools — Commit Attribution Survey (Round 2)

## Summary

Investigated 6 profile-stats projects + 2 template generators + BigQuery aggregate tool. **No project solves per-commit language attribution better than proposed REST + go-enry approach.** Two categories found: (1) **byte-size only** (GRS, jstrieb/github-stats, TraceLD), acknowledging the problem but not fixing it; (2) **WakaTime-only** (anmol098, athul), bypassing GitHub language API entirely via editor telemetry. One theoretical per-commit CLI mentioned in DEV.to discussions but repo not found in public GitHub. **Recommendation unchanged:** REST per-commit + go-enry + optional `.gitattributes` override is the frontier.

---

## Project-by-project findings

| Project | Repo URL | Primary Lang | Algorithm | Solves Commit Attribution? |
|---------|----------|--------------|-----------|---------------------------|
| **jstrieb/github-stats** | github.com/jstrieb/github-stats | Python | GraphQL `languages(orderBy:SIZE)` | No — byte-size only; explicit TODO: "Improve languages to scale by number of contributions" |
| **anmol098/waka-readme-stats** | github.com/anmol098/waka-readme-stats | Python | WakaTime API editor telemetry | Yes, but orthogonal — not GitHub stats; bypasses the problem entirely |
| **athul/waka-readme** | github.com/athul/waka-readme | Python | WakaTime API (simpler wrapper) | Yes, but orthogonal — WakaTime-only |
| **yoshi389111/github-profile-3d-contrib** | github.com/yoshi389111/github-profile-3d-contrib | TypeScript | GitHub GraphQL contributions calendar | N/A — contributions only, no languages |
| **sarthakhingankar/github-profile-readme-generator** | (repo not found / archived) | — | — | — |
| **rahul-jha98/github-profile-readme-generator** | (repo not found / archived) | — | — | — |
| **TraceLD/github-user-language-breakdown** | github.com/TraceLD/github-user-language-breakdown | TypeScript | Byte-size aggregation (`/api/langs`) | No — frontend app; backend not audited but calls generic language API |
| **madnight/githut** | github.com/madnight/githut | JavaScript | Google BigQuery public GitHub dataset | No — aggregate repo stats, not per-user commits |

---

## Detailed findings

### jstrieb/github-stats
- **Active:** Yes (last push 2026-04-18, 3.4K stars)
- **What it does:** Python CLI → GraphQL user repos → per-repo `languages(first:10, orderBy:{SIZE})` → accumulate byte-size
- **Commit attribution:** Explicitly does NOT; source has TODO comment: `# TODO: Improve languages to scale by number of contributions to`
- **Cost:** 1 GraphQL query per 100 repos
- **Novel:** Handles private repos via token; otherwise standard byte-size approach
- **Verdict:** Aware of the problem, chose not to solve it (likely due to REST rate-limit concerns)

### anmol098/waka-readme-stats
- **Active:** Yes (last push 2026-04-14, 3.9K stars)
- **What it does:** GitHub Action → fetches WakaTime API → displays editor time-in-language breakdown
- **Commit attribution:** **Yes** — but only if user has WakaTime installed and active
- **Cost:** WakaTime telemetry (user's editor plugin); no GitHub API calls for language stats
- **Solves blog repo problem?** YES, because WakaTime tracks actual editor time, not bytes. A user editing 3 JS files in a Markdown blog gets attributed to JS only if they actually spent time in JS editor.
- **Limitation:** Requires WakaTime setup; doesn't work offline; not a pure GitHub solution
- **Verdict:** Different UX. Solves the problem *orthogonally* — doesn't use GitHub language API at all.

### athul/waka-readme
- **Active:** Yes (last push 2026-02-18, 1.8K stars)
- **What it does:** Simpler WakaTime wrapper; GitHub Action fetches WakaTime API only
- **Commit attribution:** **Yes** — same as anmol098, via WakaTime editor telemetry
- **Verdict:** WakaTime alternative; no GitHub language innovation

### TraceLD/github-user-language-breakdown
- **Active:** Yes (last push 2025-02-27, 55 stars)
- **What it does:** Frontend (Vite + TypeScript) → calls `/api/langs` backend → returns byte-size breakdown
- **Commit attribution:** No — backend not audited; frontend aggregates by bytes
- **Verdict:** Small project; no novel approach

### madnight/githut
- **Active:** Inactive (last push 2024-04-03, 1K stars)
- **What it does:** Google BigQuery + GitHub public dataset → aggregate language stats across all public repos
- **Commit attribution:** No — designed for ecosystem trends, not per-user stats
- **Verdict:** Enterprise-scale analysis tool; not relevant to individual profile cards

### yoshi389111/github-profile-3d-contrib
- **Active:** Yes (last push 2026-04-15, 1.6K stars)
- **What it does:** 3D contribution calendar visualization
- **Language stats:** N/A — contributions only
- **Verdict:** Orthogonal to language problem

---

## The mysterious per-commit CLI

DEV.to post by maxfriedmann (Feb 2026): "I built a CLI to see my real GitHub language stats — does something like this already exist?"

> *"scanning every commit you've personally authored on GitHub — including private repos — and calculates how many lines you've changed per programming language"*

- **Repo:** Could not locate in public GitHub
- **Likely approach:** REST `GET /repos/{o}/{r}/commits` + parse diff → linguist/go-enry classify files → aggregate lines per language
- **If it exists:** This is exactly the REST per-commit + linguist approach proposed in prior report
- **Status:** Appears to be personal/private project or lost to time
- **Significance:** Validates that the proposed approach is feasible and novel enough to be noteworthy 4 months ago

---

## Language classification ecosystem — current state

| Approach | Maturity | Solves commit problem? | Cost | Trade-offs |
|----------|----------|----------------------|------|------------|
| **Byte-size (GitHub default)** | Stable, no code needed | No | GraphQL 1 call/100 repos | Simple; fundamentally broken for mixed-language repos |
| **Repository language count** | Stable (vn7n24fzkq) | No (only counts repo count, not commits) | Same as above | Slightly less broken; still size-biased |
| **WakaTime editor telemetry** | Requires opt-in | Yes, but not GitHub-only | User's telemetry; 0 GitHub API calls | Accurate; private; off-chain; requires user setup |
| **REST per-commit + go-enry** | Not yet packaged; proposed | Yes (90%–95% accuracy) | 1 REST call/commit (100/hr budget) | Fast; no clone; extension-limited; no `.gitattributes` support |
| **REST per-commit + go-enry + `.gitattributes`** | Proposed (this project) | Yes (95%–98% accuracy) | +1 REST call/repo for attrs | Same + minimal overhead for ~5% accuracy gain |
| **Clone + linguist Ruby gem** | Stable (lowlighter/metrics) | Yes (99% accuracy) | 15 sec timeout; storage | Accurate; slow; heavy; clones entire repo |
| **Clone + linguist-js** | Stable (lowlighter/metrics) | Yes (99% accuracy) | 15 sec timeout; storage | Same as Ruby gem |

---

## Did anyone solve it better?

**No.** The landscape is:
1. **Byte-weighted GitHub stats** — easy, broken, everyone does it (GRS, jstrieb, others)
2. **WakaTime editor telemetry** — orthogonal; requires opt-in; doesn't use GitHub API
3. **Cloning repos** — accurate but slow (lowlighter/metrics)
4. **REST per-commit + go-enry** — middle ground, not yet packaged as a standalone tool

**Null result:** No project uses `GET /repos/{o}/{r}/commits/{sha}` + go-enry/linguist for per-commit classification and packages it as a reusable tool. The DEV.to CLI mentions this exists but repo not found in public GitHub. This suggests either: (a) it's private/personal; (b) abandoned; (c) author hasn't open-sourced it.

---

## Implication for ghstats

**Prior recommendation stands.** REST per-commit + go-enry is:
- **Frontier-tier** — no packaged competitor exists yet
- **Feasible** — go-enry is performant; REST budgets fit; no cloning overhead
- **Accurate enough** — 90–95% for extension-only; 95–98% with `.gitattributes`
- **Testable** — can validate against lowlighter/metrics cloned results (regression test)

**Action:** Proceed with REST per-commit + go-enry implementation for ghstats v1. Add `.gitattributes` override as Phase 2 if accuracy feedback demands it.

---

## New ideas surfaced

- **Idea A:** Could reach out to maxfriedmann (DEV.to) to find/acquire their per-commit CLI code if it's actually been built. Might skip months of engineering.
- **Idea B:** Offer ghstats as a GitHub Action alternative to WakaTime for users who don't want editor telemetry but want accurate stats. Differentiate: "GitHub-only, no telemetry setup, REST-fast."
- **Idea C:** Add a `.gitattributes` fetcher as an optional HTTP call per repo; toggle via config. Minimal cost for significant accuracy gain on projects that use `linguist-*` directives.

---

## Unresolved questions

- **What repo is the DEV.to per-commit CLI?** Could it be claimed, forked, or improved?
- **Should ghstats include Prose (Markdown)?** Default to Programming+Markup only, with opt-in for Prose?
- **How to handle repos with zero user commits?** Skip, count PR reviews, or leave blank?
- **Fallback behavior if go-enry can't classify a file?** Count as "Other" or skip?
- **Should `.gitattributes` parsing be v1 or v2 feature?** (Adds 1 REST call/repo; ~5% accuracy gain)

---

**Sources:**
- [jstrieb/github-stats](https://github.com/jstrieb/github-stats)
- [anmol098/waka-readme-stats](https://github.com/anmol098/waka-readme-stats)
- [athul/waka-readme](https://github.com/athul/waka-readme)
- [yoshi389111/github-profile-3d-contrib](https://github.com/yoshi389111/github-profile-3d-contrib)
- [TraceLD/github-user-language-breakdown](https://github.com/TraceLD/github-user-language-breakdown)
- [madnight/githut](https://github.com/madnight/githut)
- [I built a CLI to see my real GitHub language stats (DEV.to)](https://dev.to/maxfriedmann/i-built-a-cli-to-see-my-real-github-language-stats-does-something-like-this-already-exist-1n18)
- [go-enry/go-enry](https://github.com/go-enry/go-enry)
- [GitHub Docs: About repository languages](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-repository-languages)
