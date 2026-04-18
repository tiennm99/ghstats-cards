package card

import (
	"fmt"
	"strings"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type productiveCard struct{}

func (productiveCard) Filename() string { return "4-productive-time.svg" }

// Hour ticks to label on the x-axis; same set the reference project uses.
var xTickHours = [...]int{0, 6, 12, 18, 23}

func (productiveCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	const (
		width      = 500
		height     = 220
		leftAxis   = 50
		rightPad   = 25
		topPad     = 60
		chartH     = 110
		barGap     = 2
	)
	chartW := width - leftAxis - rightPad
	barW := float64(chartW-barGap*23) / 24.0

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, "Commits by Hour (last year)"))

	max := 0
	for _, v := range p.Productive {
		if v > max {
			max = v
		}
	}
	yMax := float64(max)
	if yMax == 0 {
		yMax = 1
	}
	ticks := niceTicks(yMax, 5)
	if len(ticks) > 0 {
		yMax = ticks[len(ticks)-1]
	}

	// Y-axis: vertical line + tick marks with labels.
	fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>`,
		leftAxis, topPad, leftAxis, topPad+chartH, t.Muted)
	for _, v := range ticks {
		y := topPad + chartH - int(float64(chartH)*v/yMax)
		fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>
  <text x="%d" y="%d" font-size="10" fill="%s" text-anchor="end">%s</text>`,
			leftAxis-4, y, leftAxis, y, t.Muted,
			leftAxis-6, y+3, t.Muted, escapeXML(formatTick(v)))
	}

	// X-axis: horizontal line + tick labels.
	fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>`,
		leftAxis, topPad+chartH, leftAxis+chartW, topPad+chartH, t.Muted)
	for _, h := range xTickHours {
		x := leftAxis + int(barW*float64(h)+float64(barGap*h)+barW/2)
		fmt.Fprintf(&b, `
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s"/>
  <text x="%d" y="%d" font-size="10" fill="%s" text-anchor="middle">%d</text>`,
			x, topPad+chartH, x, topPad+chartH+4, t.Muted,
			x, topPad+chartH+16, t.Muted, h)
	}

	// Bars.
	for h := 0; h < 24; h++ {
		count := p.Productive[h]
		barH := float64(chartH) * float64(count) / yMax
		x := float64(leftAxis) + barW*float64(h) + float64(barGap*h)
		y := float64(topPad+chartH) - barH
		fmt.Fprintf(&b, `
  <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="2" fill="%s"><title>%02d:00 — %d commits</title></rect>`,
			x, y, barW, barH, t.Accent, h, count)
	}

	// X-axis caption.
	fmt.Fprintf(&b, `
  <text x="%d" y="%d" font-size="11" fill="%s" text-anchor="middle">hour of day</text>`,
		leftAxis+chartW/2, topPad+chartH+34, t.Muted)

	b.WriteString(footer)
	return []byte(b.String()), nil
}
