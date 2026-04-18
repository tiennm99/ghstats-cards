// ghstats generates SVG cards summarizing a GitHub user's profile.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tiennm99/ghstats/internal/card"
	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

func main() {
	var (
		user           = flag.String("user", "", "GitHub username (required)")
		token          = flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub token (or env GITHUB_TOKEN)")
		out            = flag.String("out", "output", "output directory")
		themesFlag     = flag.String("themes", "dracula", "comma-separated theme ids, or 'all'")
		tzName         = flag.String("tz", "Local", "timezone for productive-time card (IANA name, e.g. Asia/Saigon)")
		topRepos       = flag.Int("top-repos", 0, "optional cap on seed repos probed for commit history (0 = unlimited)")
		perRepo        = flag.Int("commits-per-repo", 500, "max commits sampled per repo (covers both last-year and all-time aggregates)")
		includeForks   = flag.Bool("include-forks", true, "include forked repos in stats and commit probing")
		includePrivate = flag.Bool("include-private", true, "include private repos (requires PAT with repo scope; silently no-op otherwise)")
		timeout        = flag.Duration("timeout", 30*time.Minute, "overall deadline for fetch phase (0 = no limit)")
		listThemes     = flag.Bool("list-themes", false, "print available theme ids and exit")
	)
	flag.Parse()

	if *listThemes {
		for _, id := range theme.IDs() {
			fmt.Println(id)
		}
		return
	}

	if *user == "" {
		fmt.Fprintln(os.Stderr, "error: -user is required")
		flag.Usage()
		os.Exit(2)
	}

	selected, err := resolveThemes(*themesFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	loc, err := time.LoadLocation(*tzName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: unknown timezone %q, falling back to UTC\n", *tzName)
		loc = time.UTC
	}

	opts := github.FetchOptions{
		IncludeForks:   *includeForks,
		IncludePrivate: *includePrivate,
	}

	// Overall fetch budget. Ctrl-C cancels in-flight HTTP requests cleanly.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if *timeout > 0 {
		var cancelTimeout context.CancelFunc
		ctx, cancelTimeout = context.WithTimeout(ctx, *timeout)
		defer cancelTimeout()
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	client := github.NewClient(*token)
	profile, err := client.FetchProfile(ctx, *user, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: fetch profile: %v\n", err)
		os.Exit(1)
	}
	profile.UTCOffsetLabel = utcOffsetLabel(loc)

	// Year-loop fetch populates SeedRepos from commitContributionsByRepository
	// plus the all-time contribution calendar; must precede FetchProductive so
	// commit-history probes land on repos where the user actually committed.
	if len(profile.ContributionYears) > 0 {
		if err := client.FetchContributionsAllTime(ctx, profile, opts); err != nil {
			fmt.Fprintf(os.Stderr, "warn: all-time contributions fetch: %v\n", err)
		}
	}

	if profile.ID != "" && len(profile.SeedRepos) > 0 {
		repos := profile.SeedRepos
		if *topRepos > 0 && len(repos) > *topRepos {
			repos = repos[:*topRepos]
		}
		if err := client.FetchProductive(ctx, profile, repos, loc, *perRepo); err != nil {
			fmt.Fprintf(os.Stderr, "warn: productive-time + commits-per-language fetch: %v\n", err)
		}
	}

	for _, t := range selected {
		if err := card.RenderAll(profile, t, *out); err != nil {
			fmt.Fprintf(os.Stderr, "error: render %s: %v\n", t.ID, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s/%s/\n", *out, t.ID)
	}
}

// utcOffsetLabel formats the location's current offset from UTC as "UTC±N.NN"
// (two-decimal hours) so half-hour zones like India (UTC+5.30) or Nepal
// (UTC+5.75) render cleanly. Matches github-profile-summary-cards' style.
func utcOffsetLabel(loc *time.Location) string {
	_, offsetSec := time.Now().In(loc).Zone()
	hours := float64(offsetSec) / 3600.0
	return fmt.Sprintf("UTC%+.2f", hours)
}

func resolveThemes(spec string) ([]theme.Theme, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("no themes specified")
	}
	if spec == "all" {
		ids := theme.IDs()
		out := make([]theme.Theme, 0, len(ids))
		for _, id := range ids {
			if t, ok := theme.Lookup(id); ok {
				out = append(out, t)
			}
		}
		return out, nil
	}
	var out []theme.Theme
	for _, id := range strings.Split(spec, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		t, ok := theme.Lookup(id)
		if !ok {
			return nil, fmt.Errorf("unknown theme %q (use -list-themes)", id)
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid themes")
	}
	return out, nil
}
