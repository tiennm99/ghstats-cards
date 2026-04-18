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
![profile](./output/dracula/0-profile-details.svg)
![repos-per-language](./output/dracula/1-repos-per-language.svg)
![most-commit-language](./output/dracula/2-most-commit-language.svg)
![stats](./output/dracula/3-stats.svg)
![productive-time](./output/dracula/4-productive-time.svg)
![contributions](./output/dracula/5-contributions.svg)
![most-commit-language-all-time](./output/dracula/6-most-commit-language-all-time.svg)
![productive-time-all-time](./output/dracula/7-productive-time-all-time.svg)
![contributions-all-time](./output/dracula/8-contributions-all-time.svg)
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

1. Ensure `go vet ./...` and `go test ./...` pass on `main`.
2. Tag: `git tag -a v1.2.0 -m "..." && git push origin v1.2.0`.
3. `release.yml` handles GHCR push + binary artifacts automatically.
4. Update any public Actions marketplace metadata if the major version changed.

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

## Troubleshooting

| Symptom | Check |
| --- | --- |
| "error: fetch profile: graphql: Could not resolve to a User" | Username typo |
| "http 401" | Token expired or lacks `read:user` |
| "http 403: rate limit exceeded" | PAT scope too narrow; token quota consumed by another workflow |
| Blank contribution chart | User has 0 contributions in their window; expected |
| Private repo data missing | `-include-private=true` not set, or PAT lacks `repo` |
| Nothing committed by the Action | Check `permissions: contents: write` in the workflow |
