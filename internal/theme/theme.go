// Package theme defines SVG color palettes used by card renderers.
//
// Palette data is ported from github-profile-summary-cards so visual output
// stays consistent with that project.
package theme

import "sort"

// Theme describes the colors applied to a rendered card.
type Theme struct {
	ID            string
	Title         string  // card title text color
	Text          string  // body text color
	Background    string  // card background fill
	Stroke        string  // card outline stroke
	StrokeOpacity float64 // 0–1; 0 hides the outline
	Muted         string  // axis / legend / secondary text
	Accent        string  // bar, donut slice fallback, tick highlight
}

// Built-in palettes ported verbatim from
// github-profile-summary-cards/src/const/theme.ts.
var themes = map[string]Theme{
	"holi":                 {Title: "#5ea9eb", Text: "#d6e7ff", Background: "#030314", Stroke: "#d6e7ff", StrokeOpacity: 1, Muted: "#5090cb", Accent: "#5090cb"},
	"2077":                 {Title: "#ff0055", Text: "#03d8f3", Background: "#141321", Stroke: "#141321", StrokeOpacity: 1, Muted: "#fcee0c", Accent: "#00ffc8"},
	"algolia":              {Title: "#00aeff", Text: "#ffffff", Background: "#050f2c", Stroke: "#000000", StrokeOpacity: 0, Muted: "#2dde98", Accent: "#00aeff"},
	"apprentice":           {Title: "#ffffff", Text: "#bcbcbc", Background: "#262626", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffffaf", Accent: "#ffffff"},
	"aura_dark":            {Title: "#ff7372", Text: "#dbdbdb", Background: "#252334", Stroke: "#000000", StrokeOpacity: 0, Muted: "#6cffd0", Accent: "#ff7372"},
	"aura":                 {Title: "#a277ff", Text: "#61ffca", Background: "#15141b", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffca85", Accent: "#a277ff"},
	"ayu_mirage":           {Title: "#f4cd7c", Text: "#c7c8c2", Background: "#1f2430", Stroke: "#000000", StrokeOpacity: 0, Muted: "#73d0ff", Accent: "#f4cd7c"},
	"bear":                 {Title: "#e03c8a", Text: "#bcb28d", Background: "#1f2023", Stroke: "#000000", StrokeOpacity: 0, Muted: "#00aeff", Accent: "#e03c8a"},
	"blue_green":           {Title: "#2f97c1", Text: "#0cf574", Background: "#040f0f", Stroke: "#000000", StrokeOpacity: 0, Muted: "#f5b700", Accent: "#2f97c1"},
	"blueberry":            {Title: "#82aaff", Text: "#27e8a7", Background: "#242938", Stroke: "#000000", StrokeOpacity: 0, Muted: "#89ddff", Accent: "#82aaff"},
	"buefy":                {Title: "#7957d5", Text: "#363636", Background: "#ffffff", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ff3860", Accent: "#7957d5"},
	"calm":                 {Title: "#e07a5f", Text: "#ebcfb2", Background: "#373f51", Stroke: "#000000", StrokeOpacity: 0, Muted: "#edae49", Accent: "#e07a5f"},
	"chartreuse_dark":      {Title: "#7fff00", Text: "#ffffff", Background: "#000000", Stroke: "#000000", StrokeOpacity: 1, Muted: "#00aeff", Accent: "#7fff00"},
	"city_lights":          {Title: "#5d8cb3", Text: "#718ca1", Background: "#1d252c", Stroke: "#000000", StrokeOpacity: 0, Muted: "#4798ff", Accent: "#5d8cb3"},
	"cobalt":               {Title: "#e683d9", Text: "#75eeb2", Background: "#193549", Stroke: "#000000", StrokeOpacity: 0, Muted: "#0480ef", Accent: "#e683d9"},
	"cobalt2":              {Title: "#ffc600", Text: "#0088ff", Background: "#193549", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffffff", Accent: "#ffc600"},
	"codeSTACKr":           {Title: "#ff652f", Text: "#ffffff", Background: "#09131b", Stroke: "#0c1a25", StrokeOpacity: 1, Muted: "#ffe400", Accent: "#ff652f"},
	"darcula":              {Title: "#ba5f17", Text: "#bebebe", Background: "#242424", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffb74d", Accent: "#ba5f17"},
	"dark":                 {Title: "#ffffff", Text: "#9f9f9f", Background: "#151515", Stroke: "#000000", StrokeOpacity: 0, Muted: "#79ff97", Accent: "#ffffff"},
	"date_night":           {Title: "#da7885", Text: "#e1b2a2", Background: "#170f0c", Stroke: "#170f0c", StrokeOpacity: 1, Muted: "#bb8470", Accent: "#da7885"},
	"default":              {Title: "#586e75", Text: "#586e75", Background: "#ffffff", Stroke: "#e4e2e2", StrokeOpacity: 1, Muted: "#586e75", Accent: "#586e75"},
	"discord_old_blurple":  {Title: "#7289da", Text: "#ffffff", Background: "#2c2f33", Stroke: "#000000", StrokeOpacity: 0, Muted: "#7289da", Accent: "#7289da"},
	"dracula":              {Title: "#ff79c6", Text: "#ffb86c", Background: "#282a36", Stroke: "#282a36", StrokeOpacity: 1, Muted: "#6272a4", Accent: "#bd93f9"},
	"flag_india":           {Title: "#ff8f1c", Text: "#509e2f", Background: "#ffffff", Stroke: "#000000", StrokeOpacity: 0, Muted: "#250e62", Accent: "#ff8f1c"},
	"github_dark":          {Title: "#0366d6", Text: "#77909c", Background: "#0d1117", Stroke: "#2e343b", StrokeOpacity: 1, Muted: "#8b949e", Accent: "#40c463"},
	"github":               {Title: "#0366d6", Text: "#586069", Background: "#ffffff", Stroke: "#e4e2e2", StrokeOpacity: 1, Muted: "#586069", Accent: "#40c463"},
	"gotham":               {Title: "#2aa889", Text: "#99d1ce", Background: "#0c1014", Stroke: "#000000", StrokeOpacity: 1, Muted: "#599cab", Accent: "#2aa889"},
	"graywhite":            {Title: "#24292e", Text: "#24292e", Background: "#ffffff", Stroke: "#000000", StrokeOpacity: 0, Muted: "#24292e", Accent: "#24292e"},
	"great_gatsby":         {Title: "#ffa726", Text: "#ffd95b", Background: "#000000", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffb74d", Accent: "#ffa726"},
	"gruvbox":              {Title: "#fabd2f", Text: "#8ec07c", Background: "#282828", Stroke: "#282828", StrokeOpacity: 1, Muted: "#fe8019", Accent: "#fe8019"},
	"highcontrast":         {Title: "#e7f216", Text: "#ffffff", Background: "#000000", Stroke: "#000000", StrokeOpacity: 0, Muted: "#00ffff", Accent: "#e7f216"},
	"jolly":                {Title: "#ff64da", Text: "#ffffff", Background: "#291b3e", Stroke: "#000000", StrokeOpacity: 0, Muted: "#a960ff", Accent: "#ff64da"},
	"kacho_ga":             {Title: "#bf4a3f", Text: "#d9c8a9", Background: "#402b23", Stroke: "#000000", StrokeOpacity: 0, Muted: "#a64833", Accent: "#bf4a3f"},
	"maroongold":           {Title: "#f7ef8a", Text: "#e0aa3e", Background: "#260000", Stroke: "#000000", StrokeOpacity: 0, Muted: "#f7ef8a", Accent: "#f7ef8a"},
	"material_palenight":   {Title: "#c792ea", Text: "#a6accd", Background: "#292d3e", Stroke: "#000000", StrokeOpacity: 0, Muted: "#89ddff", Accent: "#c792ea"},
	"merko":                {Title: "#abd200", Text: "#68b587", Background: "#0a0f0b", Stroke: "#000000", StrokeOpacity: 0, Muted: "#b7d364", Accent: "#abd200"},
	"midnight_purple":      {Title: "#9745f5", Text: "#ffffff", Background: "#000000", Stroke: "#000000", StrokeOpacity: 0, Muted: "#9f4bff", Accent: "#9745f5"},
	"moltack":              {Title: "#86092c", Text: "#574038", Background: "#f5e1c0", Stroke: "#000000", StrokeOpacity: 0, Muted: "#86092c", Accent: "#86092c"},
	"monokai":              {Title: "#eb1f6a", Text: "#ffffff", Background: "#2c292d", Stroke: "#2c292d", StrokeOpacity: 1, Muted: "#e28905", Accent: "#ae81ff"},
	"moonlight":            {Title: "#ff757f", Text: "#f8f8f8", Background: "#222436", Stroke: "#222436", StrokeOpacity: 1, Muted: "#599dff", Accent: "#ff757f"},
	"nightowl":             {Title: "#c792ea", Text: "#7fdbca", Background: "#011627", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffeb95", Accent: "#c792ea"},
	"noctis_minimus":       {Title: "#d3b692", Text: "#c5cdd3", Background: "#1b2932", Stroke: "#000000", StrokeOpacity: 0, Muted: "#72b7c0", Accent: "#d3b692"},
	"nord_bright":          {Title: "#3b4252", Text: "#2e3440", Background: "#eceff4", Stroke: "#e5e9f0", StrokeOpacity: 1, Muted: "#8fbcbb", Accent: "#88c0d0"},
	"nord_dark":            {Title: "#eceff4", Text: "#e5e9f0", Background: "#2e3440", Stroke: "#eceff4", StrokeOpacity: 1, Muted: "#8fbcbb", Accent: "#88c0d0"},
	"ocean_dark":           {Title: "#8957b2", Text: "#92d534", Background: "#151a28", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffffff", Accent: "#8957b2"},
	"omni":                 {Title: "#ff79c6", Text: "#e1e1e6", Background: "#191622", Stroke: "#000000", StrokeOpacity: 0, Muted: "#e7de79", Accent: "#ff79c6"},
	"onedark":              {Title: "#e4bf7a", Text: "#df6d74", Background: "#282c34", Stroke: "#000000", StrokeOpacity: 0, Muted: "#8eb573", Accent: "#e4bf7a"},
	"outrun":               {Title: "#ffcc00", Text: "#8080ff", Background: "#141439", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ff1aff", Accent: "#ffcc00"},
	"panda":                {Title: "#19f9d899", Text: "#ff75b5", Background: "#31353a", Stroke: "#000000", StrokeOpacity: 0, Muted: "#19f9d899", Accent: "#19f9d899"},
	"prussian":             {Title: "#bddfff", Text: "#6e93b5", Background: "#172f45", Stroke: "#000000", StrokeOpacity: 0, Muted: "#38a0ff", Accent: "#bddfff"},
	"radical":              {Title: "#fe428e", Text: "#a9fef7", Background: "#141321", Stroke: "#141321", StrokeOpacity: 1, Muted: "#f8d847", Accent: "#ae81ff"},
	"react":                {Title: "#61dafb", Text: "#ffffff", Background: "#20232a", Stroke: "#000000", StrokeOpacity: 0, Muted: "#61dafb", Accent: "#61dafb"},
	"rose_pine":            {Title: "#9ccfd8", Text: "#e0def4", Background: "#191724", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ebbcba", Accent: "#9ccfd8"},
	"shades_of_purple":     {Title: "#fad000", Text: "#a599e9", Background: "#2d2b55", Stroke: "#000000", StrokeOpacity: 0, Muted: "#b362ff", Accent: "#fad000"},
	"slateorange":          {Title: "#faa627", Text: "#ffffff", Background: "#36393f", Stroke: "#000000", StrokeOpacity: 0, Muted: "#faa627", Accent: "#faa627"},
	"solarized_dark":       {Title: "#268bd2", Text: "#839496", Background: "#073642", Stroke: "#073642", StrokeOpacity: 1, Muted: "#b58900", Accent: "#859900"},
	"solarized":            {Title: "#268bd2", Text: "#586e75", Background: "#fdf6e3", Stroke: "#fdf6e3", StrokeOpacity: 1, Muted: "#b58900", Accent: "#859900"},
	"swift":                {Title: "#000000", Text: "#000000", Background: "#f7f7f7", Stroke: "#000000", StrokeOpacity: 0, Muted: "#f05237", Accent: "#000000"},
	"synthwave":            {Title: "#e2e9ec", Text: "#e5289e", Background: "#2b213a", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ef8539", Accent: "#e2e9ec"},
	"tokyonight":           {Title: "#70a5fd", Text: "#38bdae", Background: "#1a1b27", Stroke: "#1a1b27", StrokeOpacity: 1, Muted: "#bf91f3", Accent: "#bf91f3"},
	"transparent":          {Title: "#006AFF", Text: "#417E87", Background: "#00000000", Stroke: "#000000", StrokeOpacity: 0, Muted: "#0579C3", Accent: "#006AFF"},
	"vision_friendly_dark": {Title: "#ffb000", Text: "#ffffff", Background: "#000000", Stroke: "#000000", StrokeOpacity: 0, Muted: "#785ef0", Accent: "#ffb000"},
	"vue":                  {Title: "#41b883", Text: "#000000", Background: "#ffffff", Stroke: "#e4e2e2", StrokeOpacity: 1, Muted: "#41b883", Accent: "#41b883"},
	"yeblu":                {Title: "#ffff00", Text: "#ffffff", Background: "#002046", Stroke: "#000000", StrokeOpacity: 0, Muted: "#ffff00", Accent: "#ffff00"},
	"zenburn":              {Title: "#f0dfaf", Text: "#dcdccc", Background: "#3f3f3f", Stroke: "#3f3f3f", StrokeOpacity: 1, Muted: "#8cd0d3", Accent: "#7f9f7f"},
}

func init() {
	// Populate ID fields from the map key so callers can read theme.ID.
	for id, t := range themes {
		t.ID = id
		themes[id] = t
	}
}

// Lookup returns the theme with the given id.
func Lookup(id string) (Theme, bool) {
	t, ok := themes[id]
	return t, ok
}

// IDs returns every registered theme id sorted alphabetically.
func IDs() []string {
	out := make([]string, 0, len(themes))
	for id := range themes {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
