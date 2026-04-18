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
	TotalStars         int
	TotalForks         int
	TotalCommits       int
	TotalPRs           int
	TotalIssues        int
	TotalReviews       int
	TotalContributedTo int
	TotalContributions int // lifetime contributions from calendar + restricted

	// Count of owned repos grouped by primary language, sorted desc by Value.
	ReposByLanguage []LangStat

	// Count of commits (last year, by this user) attributed to each repo's
	// primary language, sorted desc. Populated by FetchProductive.
	CommitsByLanguage []LangStat

	// Commit counts grouped by hour-of-day (0-23) in the configured timezone.
	Productive [24]int

	// TopRepos are owned repos sorted by stargazer count desc. Populated by
	// FetchProfile and consumed by FetchProductive.
	TopRepos []RepoInfo
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
