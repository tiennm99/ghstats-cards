package github

import (
	"sort"
	"time"
)

// contributionYearGQL mirrors contributionYearQuery.
type contributionYearGQL struct {
	User *struct {
		ContributionsCollection struct {
			TotalCommitContributions int `json:"totalCommitContributions"`
			ContributionCalendar     struct {
				Weeks []struct {
					ContributionDays []struct {
						ContributionCount int    `json:"contributionCount"`
						Date              string `json:"date"`
					} `json:"contributionDays"`
				} `json:"weeks"`
			} `json:"contributionCalendar"`
		} `json:"contributionsCollection"`
	} `json:"user"`
}

// FetchContributionsAllTime iterates p.ContributionYears and issues one
// contributionsCollection query per year, concatenating the daily calendar
// into p.DailyContributionsAllTime and accumulating commit counts into
// p.TotalCommitsAllTime.
//
// Cost: one GraphQL call per active year (typically 1–10 per user).
func (c *Client) FetchContributionsAllTime(p *Profile) error {
	years := append([]int(nil), p.ContributionYears...)
	sort.Ints(years) // ascending so the concatenated series is oldest→newest

	for _, y := range years {
		from := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(y, 12, 31, 23, 59, 59, 0, time.UTC)
		// Clamp the current year's window to now so GitHub doesn't reject it.
		if now := time.Now().UTC(); to.After(now) {
			to = now
		}

		vars := map[string]any{
			"login": p.Login,
			"from":  from.Format(time.RFC3339),
			"to":    to.Format(time.RFC3339),
		}
		var resp contributionYearGQL
		if err := c.query(contributionYearQuery, vars, &resp); err != nil {
			return err
		}
		if resp.User == nil {
			continue
		}

		cc := resp.User.ContributionsCollection
		p.TotalCommitsAllTime += cc.TotalCommitContributions
		for _, w := range cc.ContributionCalendar.Weeks {
			for _, d := range w.ContributionDays {
				t, err := time.Parse("2006-01-02", d.Date)
				if err != nil {
					continue
				}
				p.DailyContributionsAllTime = append(p.DailyContributionsAllTime, DailyContribution{
					Date:  t,
					Count: d.ContributionCount,
				})
			}
		}
	}
	return nil
}
