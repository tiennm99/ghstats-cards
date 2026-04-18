package card

import (
	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type reposPerLanguageCard struct{}

func (reposPerLanguageCard) Filename() string { return "1-repos-per-language.svg" }

func (reposPerLanguageCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	return renderDonutCard("Repos Per Language", p.ReposByLanguage, t), nil
}
