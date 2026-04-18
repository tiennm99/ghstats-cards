package card

import (
	"fmt"
	"strings"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type statsCard struct{}

func (statsCard) Filename() string { return "3-stats.svg" }

func (statsCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	const (
		width  = 500
		height = 220
	)

	items := []kv{
		{"Total Stars", formatInt(p.TotalStars)},
		{"Total Commits (last year)", formatInt(p.TotalCommits)},
		{"Total PRs", formatInt(p.TotalPRs)},
		{"Total Issues", formatInt(p.TotalIssues)},
		{"Total PR Reviews", formatInt(p.TotalReviews)},
		{"Contributed to (non-fork)", formatInt(p.TotalContributedTo)},
	}

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, "Stats"))

	// 2×3 grid: columns at 25 and 265, rows every 42px starting at y=80.
	for i, it := range items {
		col := i % 2
		row := i / 2
		x := 25 + col*240
		y := 80 + row*42
		fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="12" fill="%s">%s</text>
  <text x="%d" y="%d" font-size="18" font-weight="600" fill="%s">%s</text>`,
			x, y, t.Muted, escapeXML(it.label),
			x, y+22, t.Accent, escapeXML(it.value))
	}

	b.WriteString(footer)
	return []byte(b.String()), nil
}
