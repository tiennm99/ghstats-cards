package github

import (
	"time"
)

// productiveGQL is the response shape for commitHistoryQuery.
type productiveGQL struct {
	Repository *struct {
		DefaultBranchRef *struct {
			Target *struct {
				History struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []struct {
						CommittedDate string `json:"committedDate"`
					} `json:"nodes"`
				} `json:"history"`
			} `json:"target"`
		} `json:"defaultBranchRef"`
	} `json:"repository"`
}

// scaleFactor is the fixed-point multiplier used when distributing a single
// commit across several languages by byte share. Stored in LangStat.Value
// (int64) so the existing sort + percentage math keeps working; the absolute
// magnitude is irrelevant because the card renders percentages.
const scaleFactor = 10_000

// FetchProductive paginates the default-branch commit history (authored by
// the target user) for each repo up to maxPerRepo commits, and fills two
// parallel sets of aggregates on the Profile:
//
//   - Last-year: p.Productive (24h histogram) and p.CommitsByLanguage
//   - All-time:  p.ProductiveAllTime and p.CommitsByLanguageAllTime
//
// One pagination pass populates both, so the all-time cards come at no extra
// API cost beyond the pages already required for the last-year bucket.
// loc is applied to CommittedDate so the heatmap reflects the user's tz.
func (c *Client) FetchProductive(p *Profile, repos []RepoInfo, loc *time.Location, maxPerRepo int) error {
	if loc == nil {
		loc = time.UTC
	}
	yearAgo := time.Now().AddDate(-1, 0, 0)

	lastYearLang := map[string]int64{}
	allTimeLang := map[string]int64{}
	langColor := map[string]string{}

	for _, repo := range repos {
		var cursor *string
		seen := 0
		for {
			if seen >= maxPerRepo {
				break
			}
			vars := map[string]any{
				"login":  p.Login,
				"repo":   repo.Name,
				"userId": p.ID,
			}
			if cursor != nil {
				vars["after"] = *cursor
			}

			var resp productiveGQL
			if err := c.query(commitHistoryQuery, vars, &resp); err != nil {
				return err
			}
			if resp.Repository == nil || resp.Repository.DefaultBranchRef == nil ||
				resp.Repository.DefaultBranchRef.Target == nil {
				break
			}
			h := resp.Repository.DefaultBranchRef.Target.History
			for _, n := range h.Nodes {
				t, err := time.Parse(time.RFC3339, n.CommittedDate)
				if err != nil {
					continue
				}
				tl := t.In(loc)
				p.ProductiveAllTime[tl.Hour()]++
				attributeCommit(repo, allTimeLang, langColor)
				if tl.After(yearAgo) {
					p.Productive[tl.Hour()]++
					attributeCommit(repo, lastYearLang, langColor)
				}
				seen++
			}
			if !h.PageInfo.HasNextPage {
				break
			}
			end := h.PageInfo.EndCursor
			cursor = &end
		}
	}

	p.CommitsByLanguage = sortLangStats(lastYearLang, langColor)
	p.CommitsByLanguageAllTime = sortLangStats(allTimeLang, langColor)
	return nil
}

// attributeCommit distributes a single commit across the repo's languages
// proportional to byte share. Falls back to the primary language when no
// byte breakdown is available (empty repo or linguist-free repo).
func attributeCommit(repo RepoInfo, commitsByLang map[string]int64, langColor map[string]string) {
	var total int64
	for _, l := range repo.Languages {
		total += l.Bytes
	}
	if total == 0 {
		if repo.PrimaryLanguage != "" {
			commitsByLang[repo.PrimaryLanguage] += scaleFactor
			if _, ok := langColor[repo.PrimaryLanguage]; !ok {
				langColor[repo.PrimaryLanguage] = repo.PrimaryColor
			}
		}
		return
	}
	for _, l := range repo.Languages {
		share := int64(scaleFactor) * l.Bytes / total
		if share == 0 {
			continue
		}
		commitsByLang[l.Name] += share
		if _, ok := langColor[l.Name]; !ok {
			langColor[l.Name] = l.Color
		}
	}
}
