package github

// profileQuery pulls everything needed for the profile, stats and languages
// cards in one round trip. Repo pagination is handled by the caller if the
// user owns more than 100 repos.
const profileQuery = `
query($login: String!, $after: String) {
  user(login: $login) {
    id
    login
    name
    bio
    avatarUrl
    company
    location
    websiteUrl
    createdAt
    followers { totalCount }
    following { totalCount }
    pullRequests { totalCount }
    issues { totalCount }
    repositoriesContributedTo(
      first: 1
      contributionTypes: [COMMIT, PULL_REQUEST, ISSUE, PULL_REQUEST_REVIEW]
    ) { totalCount }
    contributionsCollection {
      contributionYears
      totalCommitContributions
      totalIssueContributions
      totalPullRequestContributions
      totalPullRequestReviewContributions
      totalRepositoryContributions
      restrictedContributionsCount
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
            date
          }
        }
      }
    }
    repositories(
      first: 100
      after: $after
      ownerAffiliations: OWNER
      isFork: false
      orderBy: { field: STARGAZERS, direction: DESC }
    ) {
      totalCount
      pageInfo { hasNextPage endCursor }
      nodes {
        name
        stargazerCount
        forkCount
        primaryLanguage { name color }
        languages(first: 20, orderBy: { field: SIZE, direction: DESC }) {
          edges {
            size
            node { name color }
          }
        }
      }
    }
  }
}`

// commitHistoryQuery fetches commit timestamps in the default branch of one
// repo, filtered to commits authored by the target user. Used to build the
// productive-time heatmap.
const commitHistoryQuery = `
query($login: String!, $repo: String!, $userId: ID!, $after: String) {
  repository(owner: $login, name: $repo) {
    defaultBranchRef {
      target {
        ... on Commit {
          history(first: 100, after: $after, author: { id: $userId }) {
            pageInfo { hasNextPage endCursor }
            nodes { committedDate }
          }
        }
      }
    }
  }
}`

// contributionYearQuery fetches a single year's contribution calendar days
// plus the commit total for that year. Looped in Go over user.contributionYears
// to build the all-time contribution series and lifetime commit count.
const contributionYearQuery = `
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
      totalCommitContributions
      contributionCalendar {
        weeks {
          contributionDays {
            contributionCount
            date
          }
        }
      }
    }
  }
}`
