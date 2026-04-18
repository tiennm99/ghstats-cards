package card

import (
	"fmt"
	"math"
	"strings"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

// renderDonutCard draws a donut chart with a left-side legend. Shared by the
// repos-per-language and most-commit-language cards. Up to topN slices are
// shown; smaller slices are grouped into "Other".
func renderDonutCard(title string, stats []github.LangStat, t theme.Theme) []byte {
	const (
		width     = 500
		height    = 220
		topN      = 6
		cx        = 380  // donut centre x
		cy        = 120  // donut centre y
		outerR    = 70.0 // donut outer radius
		innerR    = 38.0 // donut hole
		legendX   = 30
		legendY0  = 70
		legendDY  = 22
		swatchSz  = 12
	)

	stats = collapseOther(stats, topN)

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Title, title))

	if len(stats) == 0 {
		fmt.Fprintf(&b, `
  <text x="25" y="90" font-size="13" fill="%s">No data available.</text>`, t.Muted)
		b.WriteString(footer)
		return []byte(b.String())
	}

	var total int64
	for _, s := range stats {
		total += s.Value
	}

	// Legend (square + name + percentage).
	for i, s := range stats {
		pct := 100 * float64(s.Value) / float64(total)
		y := legendY0 + i*legendDY
		fmt.Fprintf(&b, `
  <rect x="%d" y="%d" width="%d" height="%d" fill="%s" stroke="%s" stroke-width="1"/>
  <text x="%d" y="%d" font-size="13" fill="%s">%s %.2f%%</text>`,
			legendX, y-swatchSz+2, swatchSz, swatchSz,
			colorOrAccent(s.Color, t.Accent), t.Background,
			legendX+swatchSz+8, y, t.Text,
			escapeXML(s.Name), pct)
	}

	// Donut slices.
	start := -math.Pi / 2 // 12 o'clock start
	for _, s := range stats {
		angle := 2 * math.Pi * float64(s.Value) / float64(total)
		end := start + angle
		large := 0
		if angle > math.Pi {
			large = 1
		}
		sx, sy := polar(cx, cy, outerR, start)
		ex, ey := polar(cx, cy, outerR, end)
		isx, isy := polar(cx, cy, innerR, end)
		iex, iey := polar(cx, cy, innerR, start)
		fmt.Fprintf(&b, `
  <path d="M%.2f,%.2f A%.2f,%.2f 0 %d 1 %.2f,%.2f L%.2f,%.2f A%.2f,%.2f 0 %d 0 %.2f,%.2f Z" fill="%s" stroke="%s" stroke-width="1.5"/>`,
			sx, sy, outerR, outerR, large, ex, ey,
			isx, isy, innerR, innerR, large, iex, iey,
			colorOrAccent(s.Color, t.Accent), t.Background)
		start = end
	}

	b.WriteString(footer)
	return []byte(b.String())
}

// polar returns the cartesian coordinate at (r, angle) around (cx, cy).
// Angle is in radians, measured clockwise from 3 o'clock (standard SVG).
func polar(cx, cy float64, r, angle float64) (float64, float64) {
	return cx + r*math.Cos(angle), cy + r*math.Sin(angle)
}

// collapseOther returns the top (n-1) slices plus an "Other" row summing the
// rest. When the slice fits, it's returned as-is.
func collapseOther(in []github.LangStat, n int) []github.LangStat {
	if len(in) <= n {
		return in
	}
	out := make([]github.LangStat, 0, n)
	out = append(out, in[:n-1]...)
	var rest int64
	for _, s := range in[n-1:] {
		rest += s.Value
	}
	out = append(out, github.LangStat{Name: "Other", Value: rest})
	return out
}

func colorOrAccent(c, fallback string) string {
	if c == "" {
		return fallback
	}
	return c
}
