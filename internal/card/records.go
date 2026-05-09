package card

import (
	"fmt"
	"strings"
	"time"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

type recordsCard struct{}

func (recordsCard) Filename() string { return "records.svg" }

// SVG renders six lifetime "personal-best" records mirroring the stats card's
// row layout. Records focus on extremes (peaks, firsts, lifetime totals)
// rather than the cumulative aggregates already shown by stats.svg, so the
// two cards complement instead of duplicating each other.
func (recordsCard) SVG(p *github.Profile, t theme.Theme) ([]byte, error) {
	const (
		width    = 340
		height   = 200
		rowX     = 20
		rowY0    = 55
		rowDY    = 20
		iconSize = 12
		valueX   = 320
	)

	rows := buildRecordRows(p, time.Now())

	var b strings.Builder
	b.WriteString(header(width, height, t.Background, t.Stroke, t.StrokeOpacity, t.Title, "Records (all time)"))

	scale := float64(iconSize) / 16.0
	for i, r := range rows {
		y := rowY0 + i*rowDY
		fmt.Fprintf(&b, `
  <g transform="translate(%d,%.2f) scale(%.3f)" fill="%s">%s</g>
  <text x="%d" y="%d" font-size="%d" fill="%s">%s</text>
  <text x="%d" y="%d" font-size="%d" font-weight="600" fill="%s" text-anchor="end">%s</text>`,
			rowX, float64(y-iconSize+2), scale, t.Muted, r.icon,
			rowX+iconSize+8, y, fontBody, t.Text, escapeXML(r.label),
			valueX, y, fontBody, t.Accent, escapeXML(r.value))
	}

	b.WriteString(footer)
	return []byte(b.String()), nil
}

// recordRow mirrors statRow — icon + label + accent value rendered right-aligned.
type recordRow struct {
	icon  string
	label string
	value string
}

// buildRecordRows derives the six record values from the Profile. Empty data
// yields an em-dash placeholder so the card always renders six rows.
func buildRecordRows(p *github.Profile, now time.Time) []recordRow {
	const dash = "—"

	bestCount, bestDate := peakDay(p.DailyContributionsAllTime)
	bestDayValue := dash
	if !bestDate.IsZero() {
		bestDayValue = fmt.Sprintf("%s on %s", formatInt(bestCount), bestDate.Format("2006-01-02"))
	}

	monthCount, monthDate := peakMonth(p.DailyContributionsAllTime)
	bestMonthValue := dash
	if !monthDate.IsZero() {
		bestMonthValue = fmt.Sprintf("%s in %s", formatInt(monthCount), monthDate.Format("Jan 2006"))
	}

	firstDate := firstActiveDay(p.DailyContributionsAllTime)
	firstValue := dash
	if !firstDate.IsZero() {
		firstValue = firstDate.Format("2006-01-02")
	}

	activeValue := formatInt(activeDaysCount(p.DailyContributionsAllTime))

	ageValue := dash
	if !p.CreatedAt.IsZero() {
		ageValue = fmt.Sprintf("%.1f years", accountAgeYears(p.CreatedAt, now))
	}

	langValue := formatInt(languagesUsed(p.CommitsByLanguageAllTime))

	return []recordRow{
		{iconCommit, "Best day", bestDayValue},
		{iconStar, "Best month", bestMonthValue},
		{iconCalendar, "First contribution", firstValue},
		{iconHistory, "Active days", activeValue},
		{iconClock, "On GitHub", ageValue},
		{iconGlobe, "Languages used", langValue},
	}
}

// peakDay returns the highest single-day count and its date. Ties resolve to
// the earliest date because the loop only overwrites on strictly-greater
// counts and the input is chronological.
func peakDay(days []github.DailyContribution) (int, time.Time) {
	var maxCount int
	var maxDate time.Time
	for _, d := range days {
		if d.Date.IsZero() {
			continue
		}
		if d.Count > maxCount {
			maxCount = d.Count
			maxDate = d.Date
		}
	}
	return maxCount, maxDate
}

// peakMonth aggregates contributions by calendar month and returns the
// busiest month's total + the first-of-month date for that bucket. Ties
// resolve to the earliest month — `firstSeen` records the input index where
// each bucket was opened so we don't depend on Go map iteration order.
func peakMonth(days []github.DailyContribution) (int, time.Time) {
	if len(days) == 0 {
		return 0, time.Time{}
	}
	type monthKey struct {
		year  int
		month time.Month
	}
	totals := make(map[monthKey]int)
	firstSeen := make(map[monthKey]int)
	for i, d := range days {
		if d.Date.IsZero() {
			continue
		}
		k := monthKey{d.Date.Year(), d.Date.Month()}
		if _, ok := totals[k]; !ok {
			firstSeen[k] = i
		}
		totals[k] += d.Count
	}
	var bestKey monthKey
	bestCount := 0
	bestSeen := -1
	for k, c := range totals {
		if c > bestCount || (c == bestCount && (bestSeen < 0 || firstSeen[k] < bestSeen)) {
			bestCount = c
			bestKey = k
			bestSeen = firstSeen[k]
		}
	}
	if bestCount == 0 {
		return 0, time.Time{}
	}
	return bestCount, time.Date(bestKey.year, bestKey.month, 1, 0, 0, 0, 0, time.UTC)
}

// firstActiveDay returns the first day with Count > 0. Zero time when none.
func firstActiveDay(days []github.DailyContribution) time.Time {
	for _, d := range days {
		if d.Count > 0 {
			return d.Date
		}
	}
	return time.Time{}
}

// activeDaysCount counts days where Count > 0.
func activeDaysCount(days []github.DailyContribution) int {
	var n int
	for _, d := range days {
		if d.Count > 0 {
			n++
		}
	}
	return n
}

// accountAgeYears returns now − createdAt in fractional Julian years. The
// caller renders to one decimal place, so leap-year noise stays invisible.
func accountAgeYears(createdAt, now time.Time) float64 {
	if createdAt.IsZero() {
		return 0
	}
	hours := now.Sub(createdAt).Hours()
	if hours < 0 {
		return 0
	}
	return hours / 24.0 / 365.25
}

// languagesUsed is a thin wrapper for symmetry with the other helpers.
// Upstream guarantees the slice is already deduped per language.
func languagesUsed(stats []github.LangStat) int {
	return len(stats)
}
