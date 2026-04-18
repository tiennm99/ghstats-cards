package card

import (
	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type mostCommitLanguageCard struct{}

func (mostCommitLanguageCard) Filename() string { return "most-commit-language.svg" }

func (mostCommitLanguageCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	return renderDonutCard("Most Commit Language (last year)", p.CommitsByLanguage, t), nil
}
