package main

import (
	"strings"
	"testing"
	"time"
)

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
