# Deployment Guide

Three consumption paths: **GitHub Action**, **prebuilt binaries**, **go install**.

## 1. GitHub Action (recommended for README auto-updates)

### Workflow template

File: `.github/workflows/ghstats.yml` in your profile repo.

```yaml
name: ghstats

on:
  schedule:
    - cron: "0 0 * * *"        # daily at 00:00 UTC
  workflow_dispatch:

permissions:
  contents: write              # needed for commit_changes

jobs:
  cards:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: tiennm99/ghstats@v1
        with:
          user: ${{ github.repository_owner }}
          token: ${{ secrets.GHSTATS_TOKEN }}
          themes: dracula,github_dark,tokyonight
          tz: Asia/Saigon
          # start_of_week: monday   # optional; default sunday — rotates heatmap rows + weekday bars
          include_forks: "false"
          include_private: "false"
          commit_changes: "true"
```

### Required secrets

`GHSTATS_TOKEN`: a **classic** personal access token with at minimum:

| Scope | Needed for |
| --- | --- |
| `read:user` | Basic profile fields, contribution calendar |
| `repo` | Only if `include_private: "true"` |

Fine-grained PATs and the default `${{ github.token }}` lack the introspection scope for contribution calendars in many orgs, so a classic PAT is recommended.

Create one at <https://github.com/settings/tokens> → "Generate new token (classic)" → select `read:user` (+ `repo` if needed) → save as repo secret `GHSTATS_TOKEN`.

### Embedding in README

```md
![profile](./output/dracula/profile-details.svg)
![repos-per-language](./output/dracula/repos-per-language.svg)
![most-commit-language](./output/dracula/most-commit-language.svg)
![stats](./output/dracula/stats.svg)
![productive-time](./output/dracula/productive-time.svg)
![productive-weekday](./output/dracula/productive-weekday.svg)
![contributions](./output/dracula/contributions.svg)
![contributions-heatmap](./output/dracula/contributions-heatmap.svg)
![top-starred-repos](./output/dracula/top-starred-repos.svg)
![streak](./output/dracula/streak.svg)
![most-commit-language-all-time](./output/dracula/most-commit-language-all-time.svg)
![productive-time-all-time](./output/dracula/productive-time-all-time.svg)
![productive-weekday-all-time](./output/dracula/productive-weekday-all-time.svg)
![contributions-all-time](./output/dracula/contributions-all-time.svg)
![contributions-by-year](./output/dracula/contributions-by-year.svg)
```

The Action commits SVGs to `output/<theme>/` on the default branch. GitHub serves them from the raw URL the README references.

## 2. Prebuilt binaries

Each tag under `v*` publishes:
- Linux `amd64`, `arm64`
- macOS `amd64`, `arm64`
- Windows `amd64`

Released via `.github/workflows/release.yml` which matrixes `GOOS` × `GOARCH`, strips symbols (`-ldflags="-s -w"`), and uploads tar.gz / zip to the GitHub Release.

Install:

```sh
# Linux x86_64 example
curl -L https://github.com/tiennm99/ghstats/releases/latest/download/ghstats_linux_amd64.tar.gz \
  | tar xz
./ghstats -user YOUR_USERNAME
```

## 3. go install

```sh
go install github.com/tiennm99/ghstats@latest
```

Requires Go 1.26+. Puts the binary in `$(go env GOPATH)/bin`.

## Docker image

Published to `ghcr.io/tiennm99/ghstats:<tag>` on each `v*` release via `.github/workflows/release.yml` (buildx, multi-tag: exact version, major.minor, major, latest).

The Action itself uses a runner-built image by default (`image: Dockerfile` in `action.yml`). To switch to the pre-built image for faster cold starts, edit `action.yml`:

```yaml
runs:
  using: docker
  image: docker://ghcr.io/tiennm99/ghstats:v1
```

## Release process

1. Tag: `git tag -a v1.2.0 -m "..." && git push origin v1.2.0`.
2. `release.yml` runs `go vet` + `go test` as a gate before the docker and
   binaries jobs. If tests fail, no artifacts ship.
3. On green, GHCR push + cross-platform binary artifacts happen automatically.
4. The `update-major-tag` job force-moves the floating major tag (e.g. `v1`)
   to this release's commit after test + docker + binaries all pass.
   Consumers pinned to `tiennm99/ghstats@v1` pick up the release on their
   next Action run without a workflow edit.
5. Docker base images and third-party actions are SHA-pinned (with version
   comments) so mutable-tag changes upstream can't rewrite a released image.
6. **Marketplace:** the repo is already published on the GitHub Marketplace
   as [`ghstats-cards`](https://github.com/marketplace/actions/ghstats-cards)
   (the bare `ghstats` listing was taken). New releases inherit marketplace
   visibility automatically — no manual step per release.

## Rollback

- Revert the tag: `git push --delete origin v1.2.0`, delete GitHub release, delete GHCR tag.
- Users pinned to `@v1` keep working because the previous patch is still tagged.

## Rate limit considerations

| Scenario | GraphQL calls per run | Notes |
| --- | --- | --- |
| Typical user, defaults | 15–40 | Well under 5000 pts/hr |
| Active user (8 years, 30+ seed repos) | 40–80 | Still comfortable |
| `-include-private=true` with 100+ work repos | 80–200 | Fine for daily cron |
| Adversarial user with 500+ committed repos/year | Capped by `maxRepositories: 100` per year query | Long tail drops silently |

No REST calls today. Future `-accurate-languages` mode will push toward 1000+ REST per run; schedule that mode less frequently (weekly, not daily).

The client auto-handles rate-limit responses: on 429 or 403 with `X-RateLimit-Remaining: 0`, it sleeps up to 5 minutes (honoring `Retry-After` / `X-RateLimit-Reset`) and retries once. A reset window longer than 5 min surfaces as an error so CI can reschedule instead of burning runner time. Use the `-timeout` flag (default 30m) to cap total fetch duration; `SIGINT`/`SIGTERM` cancels in-flight requests cleanly.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| "error: fetch profile: graphql: Could not resolve to a User" | Username typo |
| "http 401" | Token expired or lacks `read:user` |
| "rate limit resets in 42m (>5m0s max wait)" | Client refused to sleep through a long window; reschedule the Action |
| "http 403" on non-rate-limit path | PAT scope too narrow |
| Blank contribution chart | User has 0 contributions in their window; expected |
| Private repo data missing | `-include-private=true` not set, or PAT lacks `repo` |
| Nothing committed by the Action | Check `permissions: contents: write` in the workflow |
