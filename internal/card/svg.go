package card

import (
	"fmt"
	"strings"
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

// header returns the opening <svg> tag + background rect + title text.
func header(width, height int, bg, stroke string, strokeOpacity float64, titleColor, title string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="'Segoe UI', Ubuntu, Sans-Serif">
  <rect x="0.5" y="0.5" width="%d" height="%d" rx="6" fill="%s" stroke="%s" stroke-opacity="%.2f"/>
  <text x="20" y="30" font-size="15" font-weight="600" fill="%s">%s</text>`,
		width, height, width, height,
		width-1, height-1, bg, stroke, strokeOpacity,
		titleColor, escapeXML(title))
}

const footer = `
</svg>`
