package card

import (
	"strings"
	"testing"
	"time"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

// dayAt is a tiny helper so fixtures stay readable.
func dayAt(y int, m time.Month, d, count int) github.DailyContribution {
	return github.DailyContribution{
		Date:  time.Date(y, m, d, 0, 0, 0, 0, time.UTC),
		Count: count,
	}
}

func TestPeakDay(t *testing.T) {
	cases := []struct {
		name      string
		days      []github.DailyContribution
		wantCount int
		wantDate  time.Time
	}{
		{"empty", nil, 0, time.Time{}},
		{"all zero", []github.DailyContribution{dayAt(2025, 1, 1, 0), dayAt(2025, 1, 2, 0)}, 0, time.Time{}},
		{"single peak", []github.DailyContribution{dayAt(2025, 1, 1, 5), dayAt(2025, 1, 2, 12), dayAt(2025, 1, 3, 3)}, 12, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"tie picks earliest", []github.DailyContribution{dayAt(2025, 1, 1, 7), dayAt(2025, 1, 2, 7)}, 7, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"skip zero-date pad", []github.DailyContribution{{}, dayAt(2025, 1, 5, 9)}, 9, time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotCount, gotDate := peakDay(c.days)
			if gotCount != c.wantCount || !gotDate.Equal(c.wantDate) {
				t.Errorf("peakDay = (%d, %v), want (%d, %v)", gotCount, gotDate, c.wantCount, c.wantDate)
			}
		})
	}
}

func TestPeakMonth(t *testing.T) {
	cases := []struct {
		name      string
		days      []github.DailyContribution
		wantCount int
		wantDate  time.Time
	}{
		{"empty", nil, 0, time.Time{}},
		{"all zero", []github.DailyContribution{dayAt(2025, 1, 1, 0)}, 0, time.Time{}},
		{
			"two months, second wins",
			[]github.DailyContribution{
				dayAt(2025, 1, 1, 5), dayAt(2025, 1, 2, 5),
				dayAt(2025, 2, 1, 100),
			},
			100, time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"tie picks earliest month",
			[]github.DailyContribution{
				dayAt(2025, 1, 5, 10),
				dayAt(2025, 2, 5, 10),
			},
			10, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"month total sums correctly",
			[]github.DailyContribution{
				dayAt(2025, 3, 1, 4), dayAt(2025, 3, 15, 6), dayAt(2025, 3, 31, 2),
				dayAt(2025, 4, 1, 11),
			},
			12, time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotCount, gotDate := peakMonth(c.days)
			if gotCount != c.wantCount || !gotDate.Equal(c.wantDate) {
				t.Errorf("peakMonth = (%d, %v), want (%d, %v)", gotCount, gotDate, c.wantCount, c.wantDate)
			}
		})
	}
}

func TestFirstActiveDay(t *testing.T) {
	cases := []struct {
		name string
		days []github.DailyContribution
		want time.Time
	}{
		{"empty", nil, time.Time{}},
		{"all zero", []github.DailyContribution{dayAt(2025, 1, 1, 0), dayAt(2025, 1, 2, 0)}, time.Time{}},
		{"first nonzero", []github.DailyContribution{dayAt(2025, 1, 1, 0), dayAt(2025, 1, 2, 3), dayAt(2025, 1, 3, 7)}, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := firstActiveDay(c.days)
			if !got.Equal(c.want) {
				t.Errorf("firstActiveDay = %v, want %v", got, c.want)
			}
		})
	}
}

func TestActiveDaysCount(t *testing.T) {
	days := []github.DailyContribution{
		dayAt(2025, 1, 1, 0),
		dayAt(2025, 1, 2, 1),
		dayAt(2025, 1, 3, 0),
		dayAt(2025, 1, 4, 99),
	}
	if got := activeDaysCount(days); got != 2 {
		t.Errorf("activeDaysCount = %d, want 2", got)
	}
	if got := activeDaysCount(nil); got != 0 {
		t.Errorf("activeDaysCount(nil) = %d, want 0", got)
	}
}

func TestAccountAgeYears(t *testing.T) {
	now := time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		createdAt time.Time
		want      float64 // tolerance 0.05
	}{
		{"zero", time.Time{}, 0},
		{"future", now.Add(48 * time.Hour), 0},
		{"two years prior", time.Date(2024, 5, 9, 0, 0, 0, 0, time.UTC), 2.0},
		{"~8.7 years", time.Date(2017, 9, 1, 0, 0, 0, 0, time.UTC), 8.7},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := accountAgeYears(c.createdAt, now)
			diff := got - c.want
			if diff < -0.05 || diff > 0.05 {
				t.Errorf("accountAgeYears = %.3f, want ~%.2f", got, c.want)
			}
		})
	}
}

func TestLanguagesUsed(t *testing.T) {
	if got := languagesUsed(nil); got != 0 {
		t.Errorf("languagesUsed(nil) = %d, want 0", got)
	}
	stats := []github.LangStat{{Name: "Go"}, {Name: "TypeScript"}, {Name: "Python"}}
	if got := languagesUsed(stats); got != 3 {
		t.Errorf("languagesUsed = %d, want 3", got)
	}
}

func TestRecordsCardSVG(t *testing.T) {
	th, _ := theme.Lookup("dracula")
	p := &github.Profile{
		CreatedAt: time.Date(2017, 9, 1, 0, 0, 0, 0, time.UTC),
		DailyContributionsAllTime: []github.DailyContribution{
			dayAt(2017, 9, 5, 1),
			dayAt(2017, 9, 6, 0),
			dayAt(2026, 4, 18, 88),
			dayAt(2026, 5, 1, 13),
		},
		CommitsByLanguageAllTime: []github.LangStat{{Name: "Go"}, {Name: "Rust"}},
	}

	svg, err := recordsCard{}.SVG(p, th)
	if err != nil {
		t.Fatalf("SVG err: %v", err)
	}
	out := string(svg)

	mustContain := []string{
		"Records (all time)",
		"Best day", "88 on 2026-04-18",
		"Best month", "in Apr 2026",
		"First contribution", "2017-09-05",
		"Active days",
		"On GitHub",
		"Languages used",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("missing %q in SVG", s)
		}
	}
}

func TestRecordsCardEmptyProfile(t *testing.T) {
	th, _ := theme.Lookup("dracula")
	svg, err := recordsCard{}.SVG(&github.Profile{}, th)
	if err != nil {
		t.Fatalf("empty SVG err: %v", err)
	}
	out := string(svg)
	// All six labels still present; values fall back to em-dash.
	for _, label := range []string{"Best day", "Best month", "First contribution", "Active days", "On GitHub", "Languages used"} {
		if !strings.Contains(out, label) {
			t.Errorf("empty profile missing label %q", label)
		}
	}
	if !strings.Contains(out, "—") {
		t.Error("empty profile should emit em-dash placeholder")
	}
}
