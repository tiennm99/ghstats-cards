package card

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// escapeXML replaces the five XML-significant characters so user-controlled
// strings (bio, repo names) can't break the SVG document or inject markup.
func escapeXML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

// truncate returns s clamped to at most n runes, appending "…" when the
// input was longer. Operates on runes so multi-byte strings (emoji, CJK)
// don't split mid-codepoint. Used by every row-style card to make sure a
// pathological name / company / location can't push the card's right edge
// out past the 340 px frame.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimRight(string(r[:n-1]), ".") + "…"
}

// formatInt renders n with thousands separators (e.g. 12345 → "12,345").
func formatInt(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteByte(',')
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	out := b.String()
	if neg {
		return "-" + out
	}
	return out
}

// header returns the opening <svg> tag + background rect + title text. The
// title font-size is auto-shrunk if the string wouldn't fit on one line at
// the default 15 px — productive-time and productive-weekday append the
// timezone label ("UTC+7.00") which can push the title to ~40 chars, past
// the 340 px frame at 15 px. Floor at 11 px so small zones stay readable.
func header(width, height int, bg, stroke string, strokeOpacity float64, titleColor, title string) string {
	fontSize := fitTitleFontSize(title, width)
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="'Segoe UI', Ubuntu, Sans-Serif">
  <rect x="0.5" y="0.5" width="%d" height="%d" rx="6" fill="%s" stroke="%s" stroke-opacity="%.2f"/>
  <text x="20" y="30" font-size="%d" font-weight="600" fill="%s">%s</text>`,
		width, height, width, height,
		width-1, height-1, bg, stroke, strokeOpacity,
		fontSize, titleColor, escapeXML(title))
}

// fitTitleFontSize returns the largest integer font size in
// [titleMinFont, titleMaxFont] at which the title still fits in
// width − titleLeftInset − titleRightSafety, using the same 0.6 × fontSize
// char-width estimate as the fit-the-frame test. Computed directly from
// the budget: ideal = budget / (chars × 0.6), floored to an int, clamped
// to the allowed range.
//
// Utilization for realistic dracula titles (width=340, budget=316):
//
//	"Stats" (5)                                        → 15 px (14 %)
//	"Top Starred Repos" (17)                           → 15 px (48 %)
//	"Most Commit Language (all time)" (32)             → 15 px (91 %)
//	"Commits by Hour (last year, UTC+7.00)" (37)       → 14 px (98 %)
//	"Commits by Weekday (last year, UTC+7.00)" (40)    → 13 px (99 %)
//	"Commits by Weekday (last year, UTC+12.75)" (41)   → 12 px (93 %)
func fitTitleFontSize(title string, width int) int {
	chars := utf8.RuneCountInString(title)
	if chars == 0 {
		return titleMaxFont
	}
	budget := float64(width - titleLeftInset - titleRightSafety)
	ideal := int(math.Floor(budget / (float64(chars) * titleCharRatio)))
	if ideal > titleMaxFont {
		return titleMaxFont
	}
	if ideal < titleMinFont {
		return titleMinFont
	}
	return ideal
}

// Title-sizing constants are exported to package-private so the unit test
// can reference them without duplicating magic numbers.
const (
	titleLeftInset    = 20
	titleRightSafety  = 4
	titleMinFont      = 11
	titleMaxFont      = 15
	titleCharRatio    = 0.6
)

const footer = `
</svg>`
