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

// profileRow is one labeled-by-icon line in the profile card.
type profileRow struct {
	icon  string // raw SVG path(s) from icons.go
	value string
}

func (profileCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	const (
		width    = 500
		height   = 220
		rowX     = 25
		rowY0    = 70
		rowDY    = 24
		iconSize = 14
		maxRows  = 7
	)

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, cardTitle(p)))

	rows := buildProfileRows(p)
	if len(rows) > maxRows {
		rows = rows[:maxRows]
	}

	// scale factor to fit 16x16 octicon into iconSize box
	scale := float64(iconSize) / 16.0

	for i, r := range rows {
		y := rowY0 + i*rowDY
		// icon glyph: translate to row position, scale down, fill with muted.
		fmt.Fprintf(&b, `
  <g transform="translate(%d,%.2f) scale(%.3f)" fill="%s">%s</g>
  <text x="%d" y="%d" font-size="13" fill="%s">%s</text>`,
			rowX, float64(y-iconSize+2), scale, t.Muted, r.icon,
			rowX+iconSize+10, y, t.Text, escapeXML(r.value))
	}

	b.WriteString(footer)
	return []byte(b.String()), nil
}

// cardTitle mirrors github-profile-summary-cards: "login (Name)" when name is
// set, "login" otherwise.
func cardTitle(p *github.Profile) string {
	if p.Name != "" {
		return p.Login + " (" + p.Name + ")"
	}
	return p.Login
}

func buildProfileRows(p *github.Profile) []profileRow {
	var rows []profileRow
	if p.Company != "" {
		rows = append(rows, profileRow{icon: iconCompany, value: p.Company})
	}
	if p.Location != "" {
		rows = append(rows, profileRow{icon: iconLocation, value: p.Location})
	}
	if p.Website != "" {
		rows = append(rows, profileRow{icon: iconLink, value: p.Website})
	}
	if !p.CreatedAt.IsZero() {
		rows = append(rows, profileRow{
			icon:  iconClock,
			value: fmt.Sprintf("%s (%s)", p.CreatedAt.Format("2006-01-02"), ageAgo(p.CreatedAt)),
		})
	}
	rows = append(rows, profileRow{
		icon:  iconPeople,
		value: fmt.Sprintf("%s followers · %s following", formatInt(p.Followers), formatInt(p.Following)),
	})
	rows = append(rows, profileRow{
		icon:  iconRepos,
		value: fmt.Sprintf("%s public repos", formatInt(p.PublicRepos)),
	})
	return rows
}

// ageAgo returns "N years ago" (rounded down to whole years). Falls back to
// months or days for accounts under a year old. Matches
// github-profile-summary-cards' getProfileDateJoined.
func ageAgo(since time.Time) string {
	now := time.Now()
	y := now.Year() - since.Year()
	m := int(now.Month()) - int(since.Month())
	if m < 0 || (m == 0 && now.Day() < since.Day()) {
		y--
	}
	if y > 0 {
		return fmt.Sprintf("%d %s ago", y, plural(y, "year"))
	}

	months := (now.Year()-since.Year())*12 + int(now.Month()) - int(since.Month())
	if now.Day() < since.Day() {
		months--
	}
	if months > 0 {
		return fmt.Sprintf("%d %s ago", months, plural(months, "month"))
	}

	days := int(now.Sub(since).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return fmt.Sprintf("%d %s ago", days, plural(days, "day"))
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
