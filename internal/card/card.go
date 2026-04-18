// Package card renders Profile data into SVG cards on disk.
package card

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tiennm99/ghstats/internal/github"
	"github.com/tiennm99/ghstats/internal/theme"
)

// Card renders one SVG for a Profile under the given theme.
type Card interface {
	// Filename is the on-disk basename (e.g. "profile-details.svg").
	Filename() string
	// SVG returns the rendered SVG bytes.
	SVG(p *github.Profile, t theme.Theme) ([]byte, error)
}

// allCards is the ordered list rendered by RenderAll. Filenames are plain
// kebab-case — README authors embed them by name, not by lexicographic order.
var allCards = []Card{
	profileCard{},
	reposPerLanguageCard{},
	mostCommitLanguageCard{},
	statsCard{},
	productiveCard{},
	contributionsCard{},
	mostCommitLanguageAllTimeCard{},
	productiveAllTimeCard{},
	contributionsAllTimeCard{},
}

// RenderAll writes every card into outDir/<themeID>/.
func RenderAll(p *github.Profile, t theme.Theme, outDir string) error {
	dir := filepath.Join(outDir, t.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	for _, c := range allCards {
		data, err := c.SVG(p, t)
		if err != nil {
			return fmt.Errorf("render %s: %w", c.Filename(), err)
		}
		path := filepath.Join(dir, c.Filename())
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}
