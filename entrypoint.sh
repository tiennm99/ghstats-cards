#!/bin/sh
# Entrypoint for the ghstats GitHub Action.
# Inputs are passed via INPUT_* environment variables (set by the Action runtime).
# This script translates them into ghstats CLI flags and optionally commits the
# generated SVGs back to the repository.

set -eu

user="${INPUT_USER:-}"
token="${INPUT_TOKEN:-${GITHUB_TOKEN:-}}"
out="${INPUT_OUT:-output}"
themes="${INPUT_THEMES:-dracula}"
tz="${INPUT_TZ:-UTC}"
start_of_week="${INPUT_START_OF_WEEK:-sunday}"
top_repos="${INPUT_TOP_REPOS:-0}"
commits_per_repo="${INPUT_COMMITS_PER_REPO:-500}"
include_forks="${INPUT_INCLUDE_FORKS:-true}"
include_private="${INPUT_INCLUDE_PRIVATE:-true}"
commit_changes="${INPUT_COMMIT_CHANGES:-false}"
commit_message="${INPUT_COMMIT_MESSAGE:-chore: update ghstats cards}"
commit_branch="${INPUT_COMMIT_BRANCH:-}"
author_name="${INPUT_AUTHOR_NAME:-github-actions[bot]}"
author_email="${INPUT_AUTHOR_EMAIL:-41898282+github-actions[bot]@users.noreply.github.com}"

if [ -z "$user" ]; then
  echo "::error::input 'user' is required" >&2
  exit 2
fi

mkdir -p "$out"

echo "Running ghstats for user=$user themes=$themes out=$out"
ghstats \
  -user "$user" \
  -token "$token" \
  -out "$out" \
  -themes "$themes" \
  -tz "$tz" \
  -start-of-week "$start_of_week" \
  -top-repos "$top_repos" \
  -commits-per-repo "$commits_per_repo" \
  -include-forks="$include_forks" \
  -include-private="$include_private"

if [ "$commit_changes" = "true" ]; then
  workspace="${GITHUB_WORKSPACE:-/github/workspace}"
  cd "$workspace"
  git config --global --add safe.directory "$workspace"
  git config user.name "$author_name"
  git config user.email "$author_email"

  if [ -n "$commit_branch" ]; then
    git fetch origin "$commit_branch" || true
    git checkout -B "$commit_branch"
  fi

  git add "$out"
  if git diff --cached --quiet; then
    echo "No card changes to commit."
  else
    git commit -m "$commit_message"
    git push origin HEAD
  fi
fi
