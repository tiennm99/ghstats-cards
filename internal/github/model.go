package github

import "time"

// Profile is the aggregate of data other packages render into cards.
type Profile struct {
	ID        string
	Login     string
	Name      string
	Bio       string
	AvatarURL string
	Company   string
	Location  string
	Website   string
	CreatedAt time.Time

	Followers   int
	Following   int
	PublicRepos int

	// Totals for the stats card.
	TotalStars          int
	TotalForks          int
	TotalCommits        int // last year, from contributionsCollection
	TotalCommitsAllTime int // sum across contributionYears
	TotalPRs            int
	TotalIssues         int
	TotalReviews        int
	TotalContributedTo  int
	TotalContributions  int // lifetime contributions from calendar + restricted

	// Count of owned repos grouped by primary language, sorted desc by Value.
	ReposByLanguage []LangStat

	// Count of commits (last year, by this user) attributed to each repo's
	// primary language, sorted desc. Populated by FetchProductive.
	CommitsByLanguage []LangStat

	// Commit counts grouped by hour-of-day (0-23) in the configured timezone.
	// Productive is the last-year slice; ProductiveAllTime is the lifetime
	// slice derived from the same paginated commit history so we pay for
	// pagination once.
	Productive        [24]int
	ProductiveAllTime [24]int

	// CommitsByLanguageAllTime is the lifetime counterpart of
	// CommitsByLanguage, computed from the same commit stream.
	CommitsByLanguageAllTime []LangStat

	// DailyContributions is the raw per-day contribution calendar covering
	// the most recent year. The area chart aggregates it into monthly
	// buckets; kept granular here so any downstream card can re-bin freely.
	DailyContributions []DailyContribution

	// DailyContributionsAllTime concatenates contribution calendars across
	// every year the user has been active (user.contributionYears), so the
	// all-time area chart can show history beyond the default 1-year window.
	DailyContributionsAllTime []DailyContribution

	// TopRepos are owned repos sorted by stargazer count desc. Populated by
	// FetchProfile and consumed by FetchProductive.
	TopRepos []RepoInfo

	// ContributionYears lists every calendar year the user has been active
	// on GitHub, newest first. Used by FetchContributionsAllTime to iterate
	// per-year contributionsCollection queries.
	ContributionYears []int

	// UTCOffsetLabel is the configured timezone rendered as "UTC±N.NN" for
	// display on time-based cards (e.g. "UTC+7.00" for Asia/Saigon). Filled
	// by the CLI after loading -tz.
	UTCOffsetLabel string
}

// DailyContribution is a single day in the contributions calendar.
type DailyContribution struct {
	Date  time.Time
	Count int
}

// LangStat is one row in a language breakdown card. Value is repo count or
// commit count depending on which slice holds it.
type LangStat struct {
	Name  string
	Color string
	Value int64
}

// RepoInfo is the minimal owned-repo summary used for downstream fetches.
type RepoInfo struct {
	Name            string
	PrimaryLanguage string
	PrimaryColor    string
	// Languages is the repo's language byte breakdown as reported by GitHub
	// linguist. Used to attribute each commit fractionally across all
	// languages the repo contains, rather than crediting only the primary.
	Languages []LangEdge
}

// LangEdge is a single entry in a repo's language-bytes breakdown.
type LangEdge struct {
	Name  string
	Color string
	Bytes int64
}

// repoNode is the GraphQL shape of one repository node; kept here because
// it's shared by the profile fetcher and the productive-time fetcher.
type repoNode struct {
	Name            string `json:"name"`
	StargazerCount  int    `json:"stargazerCount"`
	ForkCount       int    `json:"forkCount"`
	PrimaryLanguage *struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"primaryLanguage"`
	Languages struct {
		Edges []struct {
			Size int64 `json:"size"`
			Node struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"languages"`
}
