package card

import (
	"fmt"
	"strings"
	"time"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type contributionsHeatmapCard struct{}

func (contributionsHeatmapCard) Filename() string { return "contributions-heatmap.svg" }

func (contributionsHeatmapCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	return renderHeatmap("Contributions (last year)", p.DailyContributions, t), nil
}

// renderHeatmap draws the classic 7×N week grid. Sunday at top, Saturday at
// bottom, oldest week on the left. Cell color mixes theme.Background with
// theme.Accent in four intensity buckets so every palette inherits a usable
// heatmap without a separate color ramp in the theme schema.
//
// Geometry is sized so 53 weeks fit inside the 340 px frame:
// leftPad (22) + 53*(cellSize+cellGap)=53*6=318 → grid ends at x=340.
func renderHeatmap(title string, days []github.DailyContribution, t theme.Theme) []byte {
	const (
		width    = 340
		height   = 200
		cellSize = 5
		cellGap  = 1
		leftPad  = 22
		topPad   = 62
	)

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, title))

	if len(days) == 0 {
		fmt.Fprintf(&b, `
  <text x="25" y="100" font-size="13" fill="%s">No contribution data available.</text>`, t.Muted)
		b.WriteString(footer)
		return []byte(b.String())
	}

	cells := padToWeekGrid(days)
	weeks := len(cells) / 7

	// Determine intensity buckets from non-zero percentiles so sparse users
	// still get visible cells and prolific users don't saturate the top bucket.
	buckets := intensityThresholds(cells)
	ramp := [5]string{
		mixHex(t.Background, t.Accent, 0.00),
		mixHex(t.Background, t.Accent, 0.25),
		mixHex(t.Background, t.Accent, 0.50),
		mixHex(t.Background, t.Accent, 0.75),
		mixHex(t.Background, t.Accent, 1.00),
	}

	// Weekday labels (Mon, Wed, Fri) printed only on alternating rows to
	// avoid visual clutter; matches GitHub's own layout.
	for i, label := range [7]string{"", "Mon", "", "Wed", "", "Fri", ""} {
		if label == "" {
			continue
		}
		y := topPad + i*(cellSize+cellGap) + cellSize - 1
		fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="9" fill="%s" text-anchor="end">%s</text>`,
			leftPad-4, y, t.Muted, label)
	}

	// Month labels across the top. We print each month the first time its
	// first day appears in a week column, skipping consecutive duplicates.
	// Labels within ~20 px of the right edge are dropped so a trailing "Dec"
	// or "Apr" can't extend past the card frame.
	const monthLabelMaxX = width - 20
	lastMonth := time.Month(0)
	for w := 0; w < weeks; w++ {
		first := cells[w*7].Date
		if first.Day() > 7 {
			continue // the 1st of the month falls in an earlier week
		}
		if first.Month() == lastMonth {
			continue
		}
		lastMonth = first.Month()
		x := leftPad + w*(cellSize+cellGap)
		if x > monthLabelMaxX {
			continue
		}
		fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="9" fill="%s">%s</text>`,
			x, topPad-4, t.Muted, first.Month().String()[:3])
	}

	// Cells.
	for w := 0; w < weeks; w++ {
		for d := 0; d < 7; d++ {
			cell := cells[w*7+d]
			if cell.Date.IsZero() {
				continue // padding slot before the first real day
			}
			fill := ramp[bucketFor(cell.Count, buckets)]
			x := leftPad + w*(cellSize+cellGap)
			y := topPad + d*(cellSize+cellGap)
			fmt.Fprintf(&b, `
  <rect x="%d" y="%d" width="%d" height="%d" rx="2" fill="%s"><title>%s — %d</title></rect>`,
				x, y, cellSize, cellSize, fill,
				cell.Date.Format("2006-01-02"), cell.Count)
		}
	}

	// Legend: "Less ▢▢▢▢▢ More" at bottom right.
	legendX := width - 110
	legendY := height - 15
	fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="9" fill="%s">Less</text>`, legendX, legendY, t.Muted)
	for i, c := range ramp {
		fmt.Fprintf(&b, `
  <rect x="%d" y="%d" width="%d" height="%d" rx="2" fill="%s"/>`,
			legendX+28+i*(cellSize+2), legendY-cellSize+2, cellSize, cellSize, c)
	}
	fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="9" fill="%s">More</text>`,
		legendX+28+5*(cellSize+2)+2, legendY, t.Muted)

	b.WriteString(footer)
	return []byte(b.String())
}

// padToWeekGrid prepends zero-date slots so the returned slice is a clean
// weeks×7 grid starting on Sunday (index 0 = Sun, 6 = Sat).
func padToWeekGrid(days []github.DailyContribution) []github.DailyContribution {
	if len(days) == 0 {
		return nil
	}
	offset := int(days[0].Date.Weekday())
	grid := make([]github.DailyContribution, offset+len(days))
	copy(grid[offset:], days)
	// Round trailing remainder up to a full week so the grid is rectangular.
	if rem := len(grid) % 7; rem != 0 {
		grid = append(grid, make([]github.DailyContribution, 7-rem)...)
	}
	return grid
}

// intensityThresholds picks four cutoffs from the non-zero counts so cells
// distribute across the 5-bucket ramp. Quartile-ish without a sort cost.
func intensityThresholds(cells []github.DailyContribution) [4]int {
	var max int
	for _, c := range cells {
		if c.Count > max {
			max = c.Count
		}
	}
	if max == 0 {
		return [4]int{1, 2, 3, 4}
	}
	// Simple linear split — works well for the common case. Power users with
	// a long right tail still fall into bucket 4 without being clipped.
	return [4]int{
		1,
		max / 4,
		max / 2,
		(3 * max) / 4,
	}
}

func bucketFor(count int, thresholds [4]int) int {
	switch {
	case count <= 0:
		return 0
	case count < thresholds[1]:
		return 1
	case count < thresholds[2]:
		return 2
	case count < thresholds[3]:
		return 3
	default:
		return 4
	}
}

// mixHex blends two "#rrggbb" colors at the given ratio (0 returns a, 1 returns b).
// Non-hex or short inputs fall back to b so a misconfigured theme still renders.
func mixHex(a, b string, ratio float64) string {
	ar, ag, ab, ok := parseHex(a)
	br, bg, bb, okb := parseHex(b)
	if !ok || !okb {
		return b
	}
	r := int(float64(ar)*(1-ratio) + float64(br)*ratio)
	g := int(float64(ag)*(1-ratio) + float64(bg)*ratio)
	bl := int(float64(ab)*(1-ratio) + float64(bb)*ratio)
	return fmt.Sprintf("#%02x%02x%02x", r, g, bl)
}

func parseHex(s string) (r, g, b int, ok bool) {
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0, false
	}
	if _, err := fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, 0, 0, false
	}
	return r, g, b, true
}
