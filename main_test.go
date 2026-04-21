package main

import (
	"strings"
	"testing"
	"time"
)

// TestParseWeekday covers the common case-insensitive inputs and confirms
// an unknown value errors (main.go then falls back to Sunday with a warn).
func TestParseWeekday(t *testing.T) {
	ok := []struct {
		in   string
		want time.Weekday
	}{
		{"", time.Sunday},
		{"sunday", time.Sunday},
		{"SUNDAY", time.Sunday},
		{"Sun", time.Sunday},
		{"  monday  ", time.Monday},
		{"Mon", time.Monday},
		{"saturday", time.Saturday},
	}
	for _, c := range ok {
		got, err := parseWeekday(c.in)
		if err != nil {
			t.Errorf("parseWeekday(%q) err=%v", c.in, err)
		}
		if got != c.want {
			t.Errorf("parseWeekday(%q)=%v want %v", c.in, got, c.want)
		}
	}
	if _, err := parseWeekday("moonday"); err == nil {
		t.Error("parseWeekday(moonday): want error, got nil")
	}
}

// TestUTCOffsetLabel checks the compact format: integer hours drop the
// minutes suffix, non-zero offsets render as `UTC±H:MM`.
func TestUTCOffsetLabel(t *testing.T) {
	cases := []struct {
		zone string
		want string // must appear in the label; exact value varies by DST
	}{
		{"UTC", "UTC+0"},
		{"Asia/Saigon", "UTC+7"},
		{"Asia/Kolkata", "UTC+5:30"},  // half-hour zone
		{"Asia/Kathmandu", "UTC+5:45"}, // quarter-hour zone
	}
	for _, tc := range cases {
		loc, err := time.LoadLocation(tc.zone)
		if err != nil {
			t.Skipf("%s unavailable: %v", tc.zone, err)
		}
		got := utcOffsetLabel(loc)
		if !strings.Contains(got, tc.want) {
			t.Errorf("utcOffsetLabel(%q) = %q, want prefix %q", tc.zone, got, tc.want)
		}
	}
}
