package card

import (
	"fmt"
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

// fitTitleFontSize picks the largest title font (between 11 and 15 px) at
// which the title still fits in width − leftInset − rightSafety. Uses the
// same 0.6 × font-size char-width estimate as the fit-the-frame test.
func fitTitleFontSize(title string, width int) int {
	const (
		leftInset    = 20
		rightSafety  = 4
		minFont      = 11
		maxFont      = 15
		avgCharRatio = 0.6
	)
	budget := float64(width - leftInset - rightSafety)
	chars := utf8.RuneCountInString(title)
	if chars == 0 {
		return maxFont
	}
	for fs := maxFont; fs >= minFont; fs-- {
		if float64(chars)*float64(fs)*avgCharRatio <= budget {
			return fs
		}
	}
	return minFont
}

const footer = `
</svg>`
