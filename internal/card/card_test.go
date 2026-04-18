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
	p := &github.Profile{
		Login:       "tiennm99",
		Name:        "Minh Tien",
		Bio:         "Test & <bio>",
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
		"0-profile-details.svg",
		"1-repos-per-language.svg",
		"2-most-commit-language.svg",
		"3-stats.svg",
		"4-productive-time.svg",
	}
	for _, name := range want {
		data, err := os.ReadFile(filepath.Join(dir, "dracula", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.HasPrefix(string(data), "<svg") {
			t.Errorf("%s: missing <svg prefix", name)
		}
		if strings.Contains(string(data), "Test & <bio>") {
			t.Errorf("%s: raw XML special characters leaked through escape", name)
		}
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
