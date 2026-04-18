package github

import (
	"errors"
	"sort"
	"time"
)

// profileGQL mirrors the GraphQL response for profileQuery.
type profileGQL struct {
	User *struct {
		ID        string `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Bio       string `json:"bio"`
		AvatarURL string `json:"avatarUrl"`
		Company   string `json:"company"`
		Location  string `json:"location"`
		Website   string `json:"websiteUrl"`
		CreatedAt string `json:"createdAt"`

		Followers struct{ TotalCount int } `json:"followers"`
		Following struct{ TotalCount int } `json:"following"`

		PullRequests struct{ TotalCount int } `json:"pullRequests"`
		Issues       struct{ TotalCount int } `json:"issues"`

		RepositoriesContributedTo struct{ TotalCount int } `json:"repositoriesContributedTo"`

		ContributionsCollection struct {
			ContributionYears                   []int `json:"contributionYears"`
			TotalCommitContributions            int   `json:"totalCommitContributions"`
			TotalIssueContributions             int   `json:"totalIssueContributions"`
			TotalPullRequestContributions       int   `json:"totalPullRequestContributions"`
			TotalPullRequestReviewContributions int   `json:"totalPullRequestReviewContributions"`
			TotalRepositoryContributions        int   `json:"totalRepositoryContributions"`
			RestrictedContributionsCount        int   `json:"restrictedContributionsCount"`
			ContributionCalendar                struct {
				TotalContributions int `json:"totalContributions"`
				Weeks              []struct {
					ContributionDays []struct {
						ContributionCount int    `json:"contributionCount"`
						Date              string `json:"date"`
					} `json:"contributionDays"`
				} `json:"weeks"`
			} `json:"contributionCalendar"`
		} `json:"contributionsCollection"`

		Repositories struct {
			TotalCount int `json:"totalCount"`
			PageInfo   struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []repoNode `json:"nodes"`
		} `json:"repositories"`
	} `json:"user"`
}

// FetchProfile collects profile, stats and repos-per-language data for the
// given user. Owned repos are paginated up to 10 pages (1000 repos) as a
// safety cap.
func (c *Client) FetchProfile(login string) (*Profile, error) {
	if login == "" {
		return nil, errors.New("empty user")
	}

	p := &Profile{Login: login}
	reposPerLang := map[string]int64{}
	langColor := map[string]string{}

	var cursor *string
	const maxPages = 10
	for page := 0; page < maxPages; page++ {
		vars := map[string]any{"login": login}
		if cursor != nil {
			vars["after"] = *cursor
		}

		var resp profileGQL
		if err := c.query(profileQuery, vars, &resp); err != nil {
			return nil, err
		}
		if resp.User == nil {
			return nil, errors.New("user not found")
		}
		u := resp.User

		if page == 0 {
			p.ID = u.ID
			p.Name = u.Name
			p.Bio = u.Bio
			p.AvatarURL = u.AvatarURL
			p.Company = u.Company
			p.Location = u.Location
			p.Website = u.Website
			if t, err := time.Parse(time.RFC3339, u.CreatedAt); err == nil {
				p.CreatedAt = t
			}
			p.Followers = u.Followers.TotalCount
			p.Following = u.Following.TotalCount
			p.PublicRepos = u.Repositories.TotalCount
			p.TotalPRs = u.PullRequests.TotalCount
			p.TotalIssues = u.Issues.TotalCount
			p.TotalContributedTo = u.RepositoriesContributedTo.TotalCount

			cc := u.ContributionsCollection
			p.TotalCommits = cc.TotalCommitContributions
			p.TotalReviews = cc.TotalPullRequestReviewContributions
			p.TotalContributions = cc.ContributionCalendar.TotalContributions + cc.RestrictedContributionsCount
			p.ContributionYears = append([]int(nil), cc.ContributionYears...)

			// Flatten week → day into a linear daily series sorted by date.
			for _, w := range cc.ContributionCalendar.Weeks {
				for _, d := range w.ContributionDays {
					t, err := time.Parse("2006-01-02", d.Date)
					if err != nil {
						continue
					}
					p.DailyContributions = append(p.DailyContributions, DailyContribution{
						Date:  t,
						Count: d.ContributionCount,
					})
				}
			}
		}

		for _, r := range u.Repositories.Nodes {
			p.TotalStars += r.StargazerCount
			p.TotalForks += r.ForkCount

			info := RepoInfo{Name: r.Name}
			if r.PrimaryLanguage != nil {
				info.PrimaryLanguage = r.PrimaryLanguage.Name
				info.PrimaryColor = r.PrimaryLanguage.Color
				reposPerLang[r.PrimaryLanguage.Name]++
				if _, ok := langColor[r.PrimaryLanguage.Name]; !ok {
					langColor[r.PrimaryLanguage.Name] = r.PrimaryLanguage.Color
				}
			}
			// Carry the repo's full language-bytes breakdown so commit
			// attribution can distribute each commit across all languages
			// the repo contains, not just the primary.
			for _, e := range r.Languages.Edges {
				info.Languages = append(info.Languages, LangEdge{
					Name:  e.Node.Name,
					Color: e.Node.Color,
					Bytes: e.Size,
				})
				if _, ok := langColor[e.Node.Name]; !ok {
					langColor[e.Node.Name] = e.Node.Color
				}
			}
			p.TopRepos = append(p.TopRepos, info)
		}

		if !u.Repositories.PageInfo.HasNextPage {
			break
		}
		end := u.Repositories.PageInfo.EndCursor
		cursor = &end
	}

	p.ReposByLanguage = sortLangStats(reposPerLang, langColor)
	return p, nil
}

// sortLangStats returns a slice sorted desc by value; ties break alphabetically.
func sortLangStats(values map[string]int64, color map[string]string) []LangStat {
	out := make([]LangStat, 0, len(values))
	for name, v := range values {
		out = append(out, LangStat{Name: name, Color: color[name], Value: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Value != out[j].Value {
			return out[i].Value > out[j].Value
		}
		return out[i].Name < out[j].Name
	})
	return out
}
