package card

import (
	"fmt"
	"strings"
	"time"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type profileCard struct{}

func (profileCard) Filename() string { return "0-profile-details.svg" }

func (profileCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	const (
		width  = 500
		height = 220
	)

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, title(p)))

	// Key-value lines; skip empty fields to avoid blank rows.
	y := 75
	rows := profileRows(p)
	for _, r := range rows {
		fmt.Fprintf(&b, `
  <text x="25" y="%d" font-size="13" fill="%s">%s</text>
  <text x="140" y="%d" font-size="13" fill="%s">%s</text>`,
			y, t.Muted, escapeXML(r.label),
			y, t.Text, escapeXML(r.value))
		y += 22
	}

	b.WriteString(footer)
	return []byte(b.String()), nil
}

func title(p *github.Profile) string {
	if p.Name != "" {
		return p.Name + "'s Profile Details"
	}
	return p.Login + "'s Profile Details"
}

type kv struct{ label, value string }

func profileRows(p *github.Profile) []kv {
	rows := []kv{{"Username", "@" + p.Login}}
	if p.Name != "" {
		rows = append(rows, kv{"Name", p.Name})
	}
	if p.Company != "" {
		rows = append(rows, kv{"Company", p.Company})
	}
	if p.Location != "" {
		rows = append(rows, kv{"Location", p.Location})
	}
	if p.Website != "" {
		rows = append(rows, kv{"Website", p.Website})
	}
	if !p.CreatedAt.IsZero() {
		rows = append(rows, kv{"Joined", p.CreatedAt.Format("2006-01-02")})
		years := time.Since(p.CreatedAt).Hours() / 24 / 365
		rows = append(rows, kv{"Account age", fmt.Sprintf("%.1f years", years)})
	}
	rows = append(rows,
		kv{"Followers", formatInt(p.Followers)},
		kv{"Following", formatInt(p.Following)},
		kv{"Public repos", formatInt(p.PublicRepos)},
	)
	if len(rows) > 7 {
		rows = rows[:7]
	}
	return rows
}
