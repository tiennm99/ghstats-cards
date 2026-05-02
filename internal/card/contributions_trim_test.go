package card

import (
	"testing"
	"time"

	"github.com/tiennm99/ghstats/internal/github"
)

// GitHub's contributionCalendar is week-aligned, so a "last year" series for a
// query made on 2026-05-02 (Saturday) starts on 2025-04-27 (Sunday). Without
// trimming the chart shows 14 calendar months and the x-axis label stride
// places 04/26 and 05/26 next to each other, overlapping. Trimming should
// drop everything before the first day of last.Month minus one year.
func TestTrimToLastYearProducesCleanThirteenMonthSpan(t *testing.T) {
	start := time.Date(2025, 4, 27, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	var days []github.DailyContribution
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, github.DailyContribution{Date: d, Count: 1})
	}

	trimmed := trimToLastYear(days)
	if len(trimmed) == 0 {
		t.Fatal("trimToLastYear returned no days")
	}
	first := trimmed[0].Date
	wantFirst := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	if !first.Equal(wantFirst) {
		t.Fatalf("first day = %s, want %s", first, wantFirst)
	}
	if !trimmed[len(trimmed)-1].Date.Equal(end) {
		t.Fatalf("last day = %s, want %s", trimmed[len(trimmed)-1].Date, end)
	}

	buckets := aggregateByMonth(trimmed)
	if len(buckets) != 13 {
		t.Fatalf("bucket count = %d, want 13 (May 2025 .. May 2026)", len(buckets))
	}
	if buckets[0].Year != 2025 || buckets[0].Month != time.May {
		t.Fatalf("first bucket = %d/%d, want 2025/May", buckets[0].Year, buckets[0].Month)
	}
	if buckets[12].Year != 2026 || buckets[12].Month != time.May {
		t.Fatalf("last bucket = %d/%d, want 2026/May", buckets[12].Year, buckets[12].Month)
	}
}

func TestTrimToLastYearEmptyInput(t *testing.T) {
	if got := trimToLastYear(nil); got != nil {
		t.Fatalf("trimToLastYear(nil) = %v, want nil", got)
	}
}
