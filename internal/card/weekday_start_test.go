package card

import (
	"strings"
	"testing"
	"time"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

// TestPadToWeekGridRotatesByWeekStart checks that the leading blank pad
// matches the configured start day — with weekStart=Monday, a Thursday-first
// series needs 3 leading zero-date slots so row 0 stays Monday.
func TestPadToWeekGridRotatesByWeekStart(t *testing.T) {
	// Thursday 2025-01-02 — int(Weekday)=4.
	thu := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	days := []github.DailyContribution{{Date: thu, Count: 1}}

	cases := []struct {
		name       string
		weekStart  time.Weekday
		wantOffset int
	}{
		{"Sunday start", time.Sunday, 4},  // Thu is row 4 of Sun..Sat
		{"Monday start", time.Monday, 3},  // Thu is row 3 of Mon..Sun
		{"Thursday start", time.Thursday, 0},
		{"Friday start", time.Friday, 6},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			grid := padToWeekGrid(days, c.weekStart)
			// The first real day should land at wantOffset; slots before it
			// carry a zero Date.
			if grid[c.wantOffset].Date != thu {
				t.Fatalf("weekStart=%v: expected Thu at offset %d, got %v", c.weekStart, c.wantOffset, grid[c.wantOffset].Date)
			}
			for i := 0; i < c.wantOffset; i++ {
				if !grid[i].Date.IsZero() {
					t.Errorf("weekStart=%v: slot %d should be blank, got %v", c.weekStart, i, grid[i].Date)
				}
			}
			if len(grid)%7 != 0 {
				t.Errorf("weekStart=%v: grid len %d not a multiple of 7", c.weekStart, len(grid))
			}
		})
	}
}

// TestRenderWeekdayRespectsWeekStart asserts bar order follows weekStart.
// Saturday (6) has the peak — with Monday start, Saturday lands at position 5,
// which should carry the accent fill.
func TestRenderWeekdayRespectsWeekStart(t *testing.T) {
	th, _ := theme.Lookup("dracula")
	var data [7]int
	data[time.Saturday] = 99 // peak

	sunSVG := string(renderWeekday("t", data, time.Sunday, th))
	monSVG := string(renderWeekday("t", data, time.Monday, th))

	// Label positions: with Sunday start, "Sun" is the first label; with
	// Monday start, "Mon" is the first label. Check the order of labels
	// by scanning the <text …>LABEL</text> occurrences in the SVG.
	sunOrder := extractWeekdayLabels(sunSVG)
	wantSun := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if !equalStrings(sunOrder, wantSun) {
		t.Errorf("Sunday-start label order = %v, want %v", sunOrder, wantSun)
	}
	monOrder := extractWeekdayLabels(monSVG)
	wantMon := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	if !equalStrings(monOrder, wantMon) {
		t.Errorf("Monday-start label order = %v, want %v", monOrder, wantMon)
	}
}

// TestRenderHeatmapLabelsRespectWeekStart asserts the left-gutter weekday
// labels rotate with weekStart (rows 1,3,5 get labels).
func TestRenderHeatmapLabelsRespectWeekStart(t *testing.T) {
	th, _ := theme.Lookup("dracula")
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	var days []github.DailyContribution
	for i := 0; i < 30; i++ {
		days = append(days, github.DailyContribution{Date: base.AddDate(0, 0, i), Count: i % 3})
	}

	cases := []struct {
		name      string
		weekStart time.Weekday
		want      []string // labels expected at rows 1, 3, 5
	}{
		{"Sunday start", time.Sunday, []string{"Mon", "Wed", "Fri"}},
		{"Monday start", time.Monday, []string{"Tue", "Thu", "Sat"}},
		{"Saturday start", time.Saturday, []string{"Sun", "Tue", "Thu"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svg := string(renderHeatmap("t", days, c.weekStart, th))
			for _, label := range c.want {
				// Row labels use text-anchor="end"; pick that form so we
				// don't collide with month labels along the top.
				needle := `text-anchor="end">` + label
				if !strings.Contains(svg, needle) {
					t.Errorf("weekStart=%v: missing row label %q", c.weekStart, label)
				}
			}
		})
	}
}

// extractWeekdayLabels pulls weekday 3-letter names out of <text …>…</text>
// blocks in the order they appear. Good enough for bar-chart cards that emit
// exactly 7 weekday labels and no colliding strings.
func extractWeekdayLabels(svg string) []string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	var out []string
	rest := svg
	for {
		idx := -1
		var pick string
		for _, n := range names {
			// Match the `>Label</text>` trailer so we catch rendered labels
			// and not the `<title>` tooltip text (which uses "—").
			needle := ">" + n + "</text>"
			i := strings.Index(rest, needle)
			if i >= 0 && (idx == -1 || i < idx) {
				idx = i
				pick = n
			}
		}
		if idx == -1 {
			return out
		}
		out = append(out, pick)
		rest = rest[idx+len(pick)+len("</text>")+1:]
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
