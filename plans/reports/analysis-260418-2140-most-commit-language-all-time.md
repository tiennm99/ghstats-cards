# Why Your Most-Commit-Language (All Time) Looks Like This

Reconstructed from live GraphQL data for user `tiennm99` on 2026-04-18.

## The card says

| Rank | Language | % |
|---|---|---|
| 1 | JavaScript | 24.96 |
| 2 | Python | 22.68 |
| 3 | C# | 19.54 |
| 4 | Go | 12.65 |
| 5 | Svelte | 11.40 |
| 6 | Other | 8.77 |

## What the algorithm actually sees

Only your **top 10 starred non-fork owned repos** get probed, up to **500 commits each**. Everything else is invisible to this card.

**Your top 10 + your commits in them + linguist byte split:**

| Repo | Your commits | Primary | Linguist bytes (non-prose) |
|---|---:|---|---|
| time-mocker | 34 | C# | C# 100% |
| adventofcode | 15 | Go | Go 100% |
| export-telegram-group-members | 9 | Python | Python 100% |
| lottery-generator | 11 | Java | Java 100% |
| ghstats | 2 | Go | Go 100% |
| rplace | **44** | JavaScript | JS 56% · Svelte 43% · HTML/CSS ~0.4% |
| thptqg2016 | 9 | JavaScript | JS 72% · CSS 27% |
| go-util | 5 | Go | Go 100% |
| try-bmad | 35 | Python | **Python 87% · JS 7%** · HTML/Svelte/Groovy/CSS ~5% |
| try-claudekit | 10 | JavaScript | JS 96% |

Total commits counted: **174**.

## Per-language derivation

Each commit contributes 1 "vote" split by linguist byte share. Summed:

| Language | Vote from… | Total |
|---|---|---:|
| **JavaScript** | rplace 24.71 · try-claudekit 9.64 · thptqg2016 6.50 · try-bmad 2.57 | **43.42** |
| **Python** | try-bmad 30.47 · export-tg 9.00 | **39.47** |
| **C#** | time-mocker 34.00 | **34.00** |
| **Go** | adventofcode 15 · go-util 5 · ghstats 2 | **22.00** |
| **Svelte** | rplace 19.14 · try-bmad 0.69 | **19.83** |
| Java | lottery-generator 11.00 | 11.00 |
| CSS | thptqg2016 2.40 · small others | 2.77 |
| HTML | spread across rplace/thptqg2016/try-bmad/try-claudekit | 1.13 |
| Groovy | try-bmad 0.37 | 0.37 |

Dividing by 174 → exactly matches the card (24.96% / 22.68% / 19.54% / 12.65% / 11.40% / 8.78%).

## Why these rank where they do

| Language | Real story |
|---|---|
| **JavaScript** dominates | Three throwaway/experiment repos (rplace, try-claudekit, thptqg2016) plus a Python project that happens to ship a JS frontend (try-bmad). The 44 rplace commits alone contribute 24.7 of the 43 JS "votes" — a single weekend project is driving the #1 slot. |
| **Python** #2 | 77% of it (30.47 of 39.47) is **try-bmad**, a 35-commit scaffolding project. The 87% Python byte share there means every commit — even a README edit — gets 87% credit to Python. |
| **C#** #3 | Every single C# vote comes from **time-mocker** (34 commits, 100% C#). This is the signal with the cleanest attribution — if you committed to that repo, it was almost certainly touching C# files. |
| **Go** surprisingly low | Only 22 votes from three repos. `ghstats` itself has **2 commits** in your window because most of this session's work hasn't been pushed yet; and `adventofcode`/`go-util` are one-off exercises. Your Go day-job probably lives in private repos we can't see. |
| **Svelte** appears out of nowhere | Linguist sees ~80KB of Svelte in **rplace** next to ~103KB of JS. Each of rplace's 44 commits credits 43% to Svelte — even commits that only touched `.js` files. That's the cost of byte-weighted attribution. |

## Why it feels wrong

Four structural reasons:

1. **Top-10 cap.** Sorted by stargazers. Your actual Go/Python dev work probably lives in repos with 0 stars but many commits — they don't make the cut.
2. **Private repos invisible.** `ownerAffiliations: OWNER` + the token's scope. Your VNG Corp code does not appear.
3. **Forks excluded.** `isFork: false`. If you hack on forks of upstream projects, those contributions vanish.
4. **Byte-weighted ≠ file-touched.** rplace credits 43% to Svelte on every commit regardless of which file you edited. A commit that fixes one line in `main.js` still counts Svelte bytes.

## What would actually fix it

Only per-commit file classification does (already scoped in earlier research — `REST /commits/{sha}` + go-enry). That path would:

- Count each commit by the **files actually touched** (Svelte only if you edited `*.svelte`).
- Support opt-in to cover all repos, not just top-10.
- Recover Markdown-heavy repos that currently vanish.

Cost: ~1000 REST calls per run vs. the current ~15 GraphQL calls. A toggleable `-accurate-languages` flag still makes sense.

## Quick wins you could take without the rewrite

- **Raise `-top-repos`** to 30–50 so less-starred but heavily-committed repos enter the sample.
- **Raise `-commits-per-repo`** past 500 if you care about lifetime depth.
- **Add `-exclude-repo rplace,try-bmad,try-claudekit`** (not implemented yet) to drop known experiment repos the way github-profile-summary-cards does.

## Unresolved questions

- Do you want the `-exclude-repo` flag landed now as a short-term fix?
- Should we expose `ownerAffiliations` so contributed-to (non-owned) repos can be included?
- Is the 500-commit cap actually binding for any of your repos, or is 174 total just "everything you've pushed"?
