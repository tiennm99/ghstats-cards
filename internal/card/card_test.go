package card

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

func TestRenderAll(t *testing.T) {
	// Name is rendered in every card's title via cardTitle(). Putting
	// XML-significant chars here exercises escapeXML through the real
	// rendering pipeline, not just through the unit test below.
	p := &github.Profile{
		Login:       "tiennm99",
		Name:        `Alice & <bob> "quoted"`,
		Company:     "VNG & <Corp>",
		Followers:   12,
		Following:   7,
		PublicRepos: 42,
		TotalStars:  1234,
		ReposByLanguage: []github.LangStat{
			{Name: "Go", Color: "#00ADD8", Value: 5},
			{Name: "TypeScript", Color: "#3178c6", Value: 3},
			{Name: "Python", Color: "", Value: 2},
		},
		CommitsByLanguage: []github.LangStat{
			{Name: "Go", Color: "#00ADD8", Value: 420},
			{Name: "Python", Color: "#3572A5", Value: 150},
		},
	}
	p.Productive[9] = 3
	p.Productive[14] = 7

	th, ok := theme.Lookup("dracula")
	if !ok {
		t.Fatal("dracula theme missing")
	}
	dir := t.TempDir()
	if err := RenderAll(p, th, dir); err != nil {
		t.Fatalf("RenderAll: %v", err)
	}

	want := []string{
		"profile-details.svg",
		"repos-per-language.svg",
		"most-commit-language.svg",
		"stats.svg",
		"productive-time.svg",
		"contributions.svg",
		"most-commit-language-all-time.svg",
		"productive-time-all-time.svg",
		"contributions-all-time.svg",
	}
	for _, name := range want {
		data, err := os.ReadFile(filepath.Join(dir, "dracula", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(data)
		if !strings.HasPrefix(content, "<svg") {
			t.Errorf("%s: missing <svg prefix", name)
		}
		// The raw Name string contains `&`, `<`, `>`, `"`. None should
		// appear unescaped in the final markup. Presence of the escaped
		// forms also verifies the title actually got rendered.
		for _, leak := range []string{`Alice & <bob>`, `VNG & <Corp>`, `"quoted"`} {
			if strings.Contains(content, leak) {
				t.Errorf("%s: raw XML special characters leaked through escape (%q)", name, leak)
			}
		}
	}
}

// TestDonutSingleSlice verifies the donut renderer handles a single-slice
// case (100%) with visible geometry. Regression guard against the empty-arc
// bug where start == end degenerates the SVG A command.
func TestDonutSingleSlice(t *testing.T) {
	th, _ := theme.Lookup("dracula")
	stats := []github.LangStat{{Name: "Go", Color: "#00ADD8", Value: 100}}
	svg := string(renderDonutCard("Test", stats, th))

	if !strings.Contains(svg, "<circle") {
		t.Errorf("single-slice donut should use <circle> primitives; got:\n%s", svg)
	}
	if strings.Contains(svg, `A70.00,70.00 0 1 1 380.00,50.00`) {
		t.Error("single-slice donut still emits degenerate arc")
	}
}

// TestDonutEmpty verifies the zero-stats fallback path.
func TestDonutEmpty(t *testing.T) {
	th, _ := theme.Lookup("dracula")
	svg := string(renderDonutCard("Test", nil, th))
	if !strings.Contains(svg, "No data available") {
		t.Errorf("empty donut should render the no-data fallback; got:\n%s", svg)
	}
}

func TestFormatInt(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		12:      "12",
		999:     "999",
		1000:    "1,000",
		12345:   "12,345",
		1234567: "1,234,567",
		-12345:  "-12,345",
	}
	for in, want := range cases {
		if got := formatInt(in); got != want {
			t.Errorf("formatInt(%d)=%q want %q", in, got, want)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	got := escapeXML(`<a & "b" 'c'>`)
	want := "&lt;a &amp; &quot;b&quot; &apos;c&apos;&gt;"
	if got != want {
		t.Errorf("escapeXML=%q want %q", got, want)
	}
}
