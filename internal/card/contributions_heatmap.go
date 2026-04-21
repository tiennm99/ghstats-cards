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
	return renderHeatmap("Contributions (last year)", p.DailyContributions, p.WeekStart, t), nil
}

// renderHeatmap draws the 53-week contribution calendar as two stacked
// halves of ~27 weeks each. A single-row version has to shrink cells to
// 4×4 to fit 53 weeks inside the 340 px width; splitting the year into two
// halves lets each half be 27 weeks wide at 8×8 cells — 4× the cell area
// and distinctly more readable, while the year still reads top-to-bottom
// left-to-right. Cell color mixes theme.Background with theme.Accent in
// four intensity buckets so every palette inherits a usable heatmap.
// weekStart controls which weekday sits on row 0 (default time.Sunday).
func renderHeatmap(title string, days []github.DailyContribution, weekStart time.Weekday, t theme.Theme) []byte {
	const (
		width    = 340
		height   = 200
		cellSize = 8
		cellGap  = 1
		leftPad  = 30
		topPadA  = 45 // top half origin (month labels land at topPadA - 4)
		halfGap  = 13 // vertical space between the two halves
	)
	halfH := 7*(cellSize+cellGap) - cellGap // 7 rows of cells occupy this many px
	topPadB := topPadA + halfH + halfGap

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, title))

	if len(days) == 0 {
		fmt.Fprintf(&b, `
  <text x="25" y="100" font-size="13" fill="%s">No contribution data available.</text>`, t.Muted)
		b.WriteString(footer)
		return []byte(b.String())
	}

	cells := padToWeekGrid(days, weekStart)
	weeks := len(cells) / 7

	buckets := intensityThresholds(cells)
	ramp := [5]string{
		mixHex(t.Background, t.Accent, 0.00),
		mixHex(t.Background, t.Accent, 0.25),
		mixHex(t.Background, t.Accent, 0.50),
		mixHex(t.Background, t.Accent, 0.75),
		mixHex(t.Background, t.Accent, 1.00),
	}

	// Split the year at the week boundary nearest the middle — the top half
	// gets the ceiling so the first half holds ≥ the second when weeks is odd.
	mid := (weeks + 1) / 2
	halves := [2]struct {
		startWeek int
		endWeek   int
		topPad    int
	}{
		{0, mid, topPadA},
		{mid, weeks, topPadB},
	}

	for _, h := range halves {
		renderHeatmapHalf(&b, cells, h.startWeek, h.endWeek, h.topPad, leftPad, cellSize, cellGap, ramp, buckets, weekStart, t)
	}

	b.WriteString(footer)
	return []byte(b.String())
}

// renderHeatmapHalf draws one half of the heatmap: weekday labels on the
// left, month labels above, and the 7×(endWeek-startWeek) grid itself.
// Labels are printed on odd rows (1, 3, 5) so the 3-per-column cadence
// matches GitHub's own calendar regardless of which weekday starts the week.
func renderHeatmapHalf(b *strings.Builder, cells []github.DailyContribution, startWeek, endWeek, topPad, leftPad, cellSize, cellGap int, ramp [5]string, buckets [4]int, weekStart time.Weekday, t theme.Theme) {
	for i := 0; i < 7; i++ {
		if i%2 == 0 {
			continue
		}
		label := weekdayShort[(int(weekStart)+i)%7]
		y := topPad + i*(cellSize+cellGap) + cellSize - 1
		fmt.Fprintf(b, `
  <text x="%d" y="%d" font-size="%d" fill="%s" text-anchor="end">%s</text>`,
			leftPad-4, y, fontAxis, t.Muted, label)
	}

	// Month labels printed the first time each month's 1st-of-the-month day
	// appears in a week column within this half. Labels that would land
	// within ~20 px of the right edge are skipped.
	monthLabelMaxX := 340 - 20
	lastMonth := time.Month(0)
	for w := startWeek; w < endWeek; w++ {
		first := cells[w*7].Date
		if first.Day() > 7 || first.Month() == lastMonth {
			continue
		}
		lastMonth = first.Month()
		x := leftPad + (w-startWeek)*(cellSize+cellGap)
		if x > monthLabelMaxX {
			continue
		}
		fmt.Fprintf(b, `
  <text x="%d" y="%d" font-size="%d" fill="%s">%s</text>`,
			x, topPad-4, fontAxis, t.Muted, first.Month().String()[:3])
	}

	// Cells for this half.
	for w := startWeek; w < endWeek; w++ {
		for d := 0; d < 7; d++ {
			cell := cells[w*7+d]
			if cell.Date.IsZero() {
				continue
			}
			fill := ramp[bucketFor(cell.Count, buckets)]
			x := leftPad + (w-startWeek)*(cellSize+cellGap)
			y := topPad + d*(cellSize+cellGap)
			fmt.Fprintf(b, `
  <rect x="%d" y="%d" width="%d" height="%d" rx="1.5" fill="%s"><title>%s — %d</title></rect>`,
				x, y, cellSize, cellSize, fill,
				cell.Date.Format("2006-01-02"), cell.Count)
		}
	}
}

// padToWeekGrid prepends zero-date slots so the returned slice is a clean
// weeks×7 grid where row 0 corresponds to the configured weekStart. For
// example, with weekStart = time.Monday, a series beginning on a Thursday
// gets 3 leading blanks so row 0 stays Monday.
func padToWeekGrid(days []github.DailyContribution, weekStart time.Weekday) []github.DailyContribution {
	if len(days) == 0 {
		return nil
	}
	offset := (int(days[0].Date.Weekday()) - int(weekStart) + 7) % 7
	grid := make([]github.DailyContribution, offset+len(days))
	copy(grid[offset:], days)
	// Round trailing remainder up to a full week so the grid is rectangular.
	if rem := len(grid) % 7; rem != 0 {
		grid = append(grid, make([]github.DailyContribution, 7-rem)...)
	}
	return grid
}

// weekdayShort is the 3-letter weekday name at index = int(time.Weekday).
// Shared by the heatmap row labels and the productive-weekday bars so the
// two cards agree on spelling.
var weekdayShort = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

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
