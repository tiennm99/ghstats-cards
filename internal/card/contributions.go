package card

import (
	"fmt"
	"strings"
	"time"

	"github.com/tiennm99/ghstats-cards/internal/github"
	"github.com/tiennm99/ghstats-cards/internal/theme"
)

type contributionsCard struct{}

func (contributionsCard) Filename() string { return "contributions.svg" }

func (contributionsCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	return renderContributions("Contributions (last year)", p.DailyContributions, t), nil
}

type contributionsAllTimeCard struct{}

func (contributionsAllTimeCard) Filename() string { return "contributions-all-time.svg" }

func (contributionsAllTimeCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	return renderContributions("Contributions (all time)", p.DailyContributionsAllTime, t), nil
}

// monthBucket holds a calendar month's aggregate contribution count.
type monthBucket struct {
	Year  int
	Month time.Month
	Count int
}

// renderContributions draws a smooth filled area chart with y-axis labels on
// both sides and YY/MM labels on the x-axis. Shared by last-year and all-time
// contribution cards; only title + data differ.
func renderContributions(title string, days []github.DailyContribution, t theme.Theme) []byte {
	const (
		width    = 340
		height   = 200
		leftPad  = 28
		rightPad = 28
		topPad   = 45
		chartH   = 105
	)
	chartW := width - leftPad - rightPad

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, title))

	buckets := aggregateByMonth(days)
	if len(buckets) < 2 {
		fmt.Fprintf(&b, `
  <text x="25" y="90" font-size="13" fill="%s">No contribution data available.</text>`, t.Muted)
		b.WriteString(footer)
		return []byte(b.String())
	}

	// Y scale based on max monthly count; nice ticks for labels.
	var maxVal int
	for _, bk := range buckets {
		if bk.Count > maxVal {
			maxVal = bk.Count
		}
	}
	yMax := float64(maxVal)
	if yMax == 0 {
		yMax = 1
	}
	ticks := niceTicks(yMax, 5)
	if len(ticks) > 0 {
		yMax = ticks[len(ticks)-1]
	}

	// Map each bucket to an (x, y) point on the chart grid.
	pts := make([][2]float64, len(buckets))
	for i, bk := range buckets {
		x := float64(leftPad) + float64(chartW)*float64(i)/float64(len(buckets)-1)
		y := float64(topPad+chartH) - float64(chartH)*float64(bk.Count)/yMax
		pts[i] = [2]float64{x, y}
	}

	// Y axis mirrored on both sides with matching nice ticks.
	rightX := leftPad + chartW
	fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>`,
		leftPad, topPad, leftPad, topPad+chartH, t.Muted,
		rightX, topPad, rightX, topPad+chartH, t.Muted)
	for _, v := range ticks {
		y := topPad + chartH - int(float64(chartH)*v/yMax)
		label := escapeXML(formatTick(v))
		fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>
  <text x="%d" y="%d" font-size="10" fill="%s" text-anchor="end">%s</text>
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>
  <text x="%d" y="%d" font-size="10" fill="%s" text-anchor="start">%s</text>`,
			leftPad-4, y, leftPad, y, t.Muted,
			leftPad-6, y+3, t.Muted, label,
			rightX, y, rightX+4, y, t.Muted,
			rightX+6, y+3, t.Muted, label)
	}

	// X axis baseline + month labels. Stride is chosen so roughly xLabelTarget
	// labels span the full width regardless of bucket count — keeps the axis
	// readable whether we plot 12 months (last year) or 100+ months (all time).
	fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>`,
		leftPad, topPad+chartH, leftPad+chartW, topPad+chartH, t.Muted)
	for i, bk := range buckets {
		if !xAxisLabelVisible(i, len(buckets)) {
			continue
		}
		x := int(pts[i][0])
		label := fmt.Sprintf("%02d/%02d", int(bk.Month), bk.Year%100)
		fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>
  <text x="%d" y="%d" font-size="10" fill="%s" text-anchor="middle">%s</text>`,
			x, topPad+chartH, x, topPad+chartH+4, t.Muted,
			x, topPad+chartH+16, t.Muted, label)
	}

	// Smooth filled area using Catmull-Rom → cubic Bezier segments.
	path := catmullRomAreaPath(pts, float64(topPad+chartH))
	fmt.Fprintf(&b, `
  <path d="%s" fill="%s" fill-opacity="0.25" stroke="none"/>
  <path d="%s" fill="none" stroke="%s" stroke-width="2"/>`,
		path, t.Accent,
		catmullRomLinePath(pts), t.Accent)

	// Axis caption — documents the tick label format.
	fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="11" fill="%s" text-anchor="middle">mm/yy</text>`,
		leftPad+chartW/2, topPad+chartH+34, t.Muted)

	b.WriteString(footer)
	return []byte(b.String())
}

// xAxisLabelVisible returns true when the bucket at position i should get a
// printed month label. Targets ~xLabelTarget labels evenly distributed across
// the axis, always pinning the first and last so the span is obvious.
const xLabelTarget = 6

func xAxisLabelVisible(i, n int) bool {
	if n <= xLabelTarget {
		return true // few enough points — label all of them
	}
	if i == 0 || i == n-1 {
		return true
	}
	stride := (n - 1) / (xLabelTarget - 1)
	if stride < 1 {
		stride = 1
	}
	// Skip labels that are too close to the pinned last point.
	if n-1-i < stride/2 {
		return false
	}
	return i%stride == 0
}

// aggregateByMonth bins the daily series into consecutive month buckets
// sorted oldest→newest. Empty months between first and last are kept as
// zero-count rows so the area chart remains time-continuous.
func aggregateByMonth(days []github.DailyContribution) []monthBucket {
	if len(days) == 0 {
		return nil
	}
	counts := map[string]int{}
	for _, d := range days {
		key := fmt.Sprintf("%04d-%02d", d.Date.Year(), int(d.Date.Month()))
		counts[key] += d.Count
	}

	// Walk from the first to last month of the calendar inclusively.
	first := days[0].Date
	last := days[len(days)-1].Date
	cur := time.Date(first.Year(), first.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(last.Year(), last.Month(), 1, 0, 0, 0, 0, time.UTC)

	var out []monthBucket
	for !cur.After(end) {
		key := fmt.Sprintf("%04d-%02d", cur.Year(), int(cur.Month()))
		out = append(out, monthBucket{
			Year:  cur.Year(),
			Month: cur.Month(),
			Count: counts[key],
		})
		cur = cur.AddDate(0, 1, 0)
	}
	return out
}

// catmullRomLinePath produces an SVG path string that passes through each
// point using cubic Bezier segments derived from the Catmull-Rom spline
// (tension = 0.5, the classic d3.curveCatmullRom default).
func catmullRomLinePath(pts [][2]float64) string {
	if len(pts) < 2 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "M%.2f,%.2f", pts[0][0], pts[0][1])
	for i := 0; i < len(pts)-1; i++ {
		p0 := pts[max(i-1, 0)]
		p1 := pts[i]
		p2 := pts[i+1]
		p3 := pts[min(i+2, len(pts)-1)]

		c1x := p1[0] + (p2[0]-p0[0])/6
		c1y := p1[1] + (p2[1]-p0[1])/6
		c2x := p2[0] - (p3[0]-p1[0])/6
		c2y := p2[1] - (p3[1]-p1[1])/6

		fmt.Fprintf(&b, " C%.2f,%.2f %.2f,%.2f %.2f,%.2f",
			c1x, c1y, c2x, c2y, p2[0], p2[1])
	}
	return b.String()
}

// catmullRomAreaPath closes the smooth line path down to the baseline so it
// can be filled as an area under the curve.
func catmullRomAreaPath(pts [][2]float64, baseline float64) string {
	line := catmullRomLinePath(pts)
	if line == "" {
		return ""
	}
	return fmt.Sprintf("%s L%.2f,%.2f L%.2f,%.2f Z",
		line, pts[len(pts)-1][0], baseline, pts[0][0], baseline)
}
