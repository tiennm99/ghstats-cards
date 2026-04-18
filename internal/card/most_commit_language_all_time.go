package card

import (
	"github.com/tiennm99/ghstats-cards/internal/github"
	"github.com/tiennm99/ghstats-cards/internal/theme"
)

type mostCommitLanguageAllTimeCard struct{}

func (mostCommitLanguageAllTimeCard) Filename() string {
	return "most-commit-language-all-time.svg"
}

func (mostCommitLanguageAllTimeCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	return renderDonutCard("Most Commit Language (all time)", p.CommitsByLanguageAllTime, t), nil
}
