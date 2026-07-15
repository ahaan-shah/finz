package main

import "github.com/charmbracelet/lipgloss"

// Theme mirrors the subset of Textual's Theme dataclass this app actually
// reads (tui.py only ever touches theme.error/success/warning/accent at
// render time, plus $primary/$surface/$panel/$foreground/$text-muted via
// CSS) - every field's value below was read directly out of a running
// Textual App's resolved CSS variables (App.get_css_variables()) for each
// built-in theme, not hand-approximated, so this is the exact palette
// Textual itself would paint.
//
// Background/Surface/Panel/Foreground/FooterBackground/FooterForeground/
// FooterKeyForeground/FooterDescriptionForeground are empty for the ansi-*
// themes, matching Textual's own "ansi_default"/"transparent" values for
// them - they deliberately leave the terminal's own colors alone. Primary/
// Warning/Error/Success/Accent/FooterKeyForeground instead hold a base-16
// ANSI color number ("0"-"15") for those two themes, since Textual maps
// them to named ansi_* colors (ansi_red, ansi_green, ...) rather than a
// fixed truecolor hex.
//
// Muted approximates Textual's "$text-muted" (defined as `auto 60%`, i.e.
// auto-selected black/white composited over the background at 60%
// opacity) since Go/lipgloss has no equivalent of Rich's `auto` color -
// computed as a 60/40 blend of white-or-black (whichever `Dark` picks)
// over Background, except for the ansi-* themes, which reuse the same
// ANSI-gray convention (base color "8"/"7") splitsy already established
// for the same "no fixed hex, respect the terminal" situation.
type Theme struct {
	Name    string
	Dark    bool
	Primary lipgloss.Color // $primary - the modal form's "Save" text color
	Warning lipgloss.Color // budget gauge's 70%-100% threshold color
	Error   lipgloss.Color // $error - expenses, over-budget, delete/danger chrome
	Success lipgloss.Color // $success - income, budget gauge's under-70% color
	Accent  lipgloss.Color // $accent - borders, chart bars, budget box border

	Background lipgloss.Color
	Surface    lipgloss.Color // modal dialog background
	Panel      lipgloss.Color // header/footer bar background
	Foreground lipgloss.Color
	Muted      lipgloss.Color // $text-muted - the modal form's "Cancel" text color

	FooterBackground            lipgloss.Color
	FooterForeground            lipgloss.Color
	FooterKeyForeground         lipgloss.Color
	FooterDescriptionForeground lipgloss.Color
}

// themes is the exact built-in theme roster of the installed Textual
// version tui.py runs against (21 themes: App.available_themes as of
// textual's current release) - not a curated subset, the whole registry,
// since finz's command palette Theme page is Textual's own built-in one
// and lists every one of them.
var themes = []Theme{
	{Name: "textual-dark", Dark: true, Background: "#121212", Surface: "#1E1E1E", Panel: "#242F38", Foreground: "#E0E0E0", Primary: "#0178D4", Warning: "#FEA62B", Error: "#B93C5B", Success: "#4EBF71", Accent: "#FEA62B", Muted: "#A0A0A0", FooterBackground: "#242F38", FooterForeground: "#E0E0E0", FooterKeyForeground: "#FFA62B", FooterDescriptionForeground: "#E0E0E0"},
	{Name: "textual-light", Dark: false, Background: "#E0E0E0", Surface: "#D8D8D8", Panel: "#D0D0D0", Foreground: "#1F1F1F", Primary: "#004578", Warning: "#FEA62B", Error: "#B93C5B", Success: "#4EBF71", Accent: "#FEA62B", Muted: "#5A5A5A", FooterBackground: "#D0D0D0", FooterForeground: "#1F1F1F", FooterKeyForeground: "#0178D4", FooterDescriptionForeground: "#1F1F1F"},
	{Name: "nord", Dark: true, Background: "#2E3440", Surface: "#3B4252", Panel: "#434C5E", Foreground: "#D8DEE9", Primary: "#88C0D0", Warning: "#EACB8B", Error: "#BE616A", Success: "#A3BE8C", Accent: "#B48EAD", Muted: "#ABAEB3", FooterBackground: "#434C5E", FooterForeground: "#D8DEE9", FooterKeyForeground: "#88C0D0", FooterDescriptionForeground: "#D8DEE9"},
	{Name: "gruvbox", Dark: true, Background: "#282828", Surface: "#3C3836", Panel: "#504945", Foreground: "#FBF1C7", Primary: "#85A598", Warning: "#FD8019", Error: "#FA4934", Success: "#B7BB26", Accent: "#F9BD2F", Muted: "#A9A9A9", FooterBackground: "#504945", FooterForeground: "#FBF1C7", FooterKeyForeground: "#FABD2F", FooterDescriptionForeground: "#FBF1C7"},
	{Name: "catppuccin-mocha", Dark: true, Background: "#181825", Surface: "#313244", Panel: "#45475A", Foreground: "#CDD6F4", Primary: "#F5C2E7", Warning: "#FAE3B0", Error: "#F28FAD", Success: "#ABE9B3", Accent: "#F9B387", Muted: "#A3A3A8", FooterBackground: "#45475A", FooterForeground: "#CDD6F4", FooterKeyForeground: "#FAB387", FooterDescriptionForeground: "#CDD6F4"},
	{Name: "dracula", Dark: true, Background: "#282A36", Surface: "#2B2E3B", Panel: "#313442", Foreground: "#F8F8F2", Primary: "#BD93F9", Warning: "#FEB86C", Error: "#FE5555", Success: "#50FA7B", Accent: "#FF79C6", Muted: "#A9AAAF", FooterBackground: "#313442", FooterForeground: "#F8F8F2", FooterKeyForeground: "#FF79C6", FooterDescriptionForeground: "#F8F8F2"},
	{Name: "tokyo-night", Dark: true, Background: "#1A1B26", Surface: "#24283B", Panel: "#414868", Foreground: "#A9B1D6", Primary: "#BB9AF7", Warning: "#DFAF68", Error: "#F6768E", Success: "#9ECE6A", Accent: "#FE9E64", Muted: "#A3A4A8", FooterBackground: "#414868", FooterForeground: "#A9B1D6", FooterKeyForeground: "#FF9E64", FooterDescriptionForeground: "#A9B1D6"},
	{Name: "monokai", Dark: true, Background: "#272822", Surface: "#2E2E2E", Panel: "#3E3D32", Foreground: "#D6D6D6", Primary: "#AE81FF", Warning: "#FC971F", Error: "#F82672", Success: "#A5E22E", Accent: "#66D9EF", Muted: "#A9A9A7", FooterBackground: "#3E3D32", FooterForeground: "#D6D6D6", FooterKeyForeground: "#66D9EF", FooterDescriptionForeground: "#D6D6D6"},
	{Name: "flexoki", Dark: true, Background: "#100F0F", Surface: "#1C1B1A", Panel: "#282726", Foreground: "#FFFCF0", Primary: "#205EA6", Warning: "#AC8301", Error: "#AE3029", Success: "#65800B", Accent: "#9B76C8", Muted: "#9F9F9F", FooterBackground: "#282726", FooterForeground: "#FFFCF0", FooterKeyForeground: "#9B76C8", FooterDescriptionForeground: "#FFFCF0"},
	{Name: "catppuccin-latte", Dark: false, Background: "#EFF1F5", Surface: "#E6E9EF", Panel: "#CCD0DA", Foreground: "#4C4F69", Primary: "#8839EF", Warning: "#DE8E1D", Error: "#D10F39", Success: "#40A02B", Accent: "#FD640B", Muted: "#606062", FooterBackground: "#CCD0DA", FooterForeground: "#4C4F69", FooterKeyForeground: "#FE640B", FooterDescriptionForeground: "#4C4F69"},
	{Name: "catppuccin-frappe", Dark: true, Background: "#303446", Surface: "#414559", Panel: "#51576D", Foreground: "#C6D0F5", Primary: "#CA9EE6", Warning: "#E4C890", Error: "#E68284", Success: "#A6D189", Accent: "#F4B8E4", Muted: "#ACAEB5", FooterBackground: "#51576D", FooterForeground: "#C6D0F5", FooterKeyForeground: "#F4B8E4", FooterDescriptionForeground: "#C6D0F5"},
	{Name: "catppuccin-macchiato", Dark: true, Background: "#24273A", Surface: "#363A4F", Panel: "#494D64", Foreground: "#CAD3F5", Primary: "#C6A0F6", Warning: "#EED49F", Error: "#ED8796", Success: "#A6DA95", Accent: "#F5BDE6", Muted: "#A7A9B0", FooterBackground: "#494D64", FooterForeground: "#CAD3F5", FooterKeyForeground: "#F5BDE6", FooterDescriptionForeground: "#CAD3F5"},
	{Name: "solarized-light", Dark: false, Background: "#FDF6E3", Surface: "#EEE8D5", Panel: "#EEE8D5", Foreground: "#586E75", Primary: "#268BD2", Warning: "#CA4B16", Error: "#DB322F", Success: "#849900", Accent: "#6C71C4", Muted: "#65625B", FooterBackground: "#268BD2", FooterForeground: "#586E75", FooterKeyForeground: "#FDF6E3", FooterDescriptionForeground: "#FDF6E3"},
	{Name: "solarized-dark", Dark: true, Background: "#002B36", Surface: "#073642", Panel: "#073642", Foreground: "#839496", Primary: "#268BD2", Warning: "#CA4B16", Error: "#DB322F", Success: "#849900", Accent: "#6C71C4", Muted: "#99AAAF", FooterBackground: "#268BD2", FooterForeground: "#839496", FooterKeyForeground: "#FDF6E3", FooterDescriptionForeground: "#FDF6E3"},
	{Name: "rose-pine", Dark: true, Background: "#191724", Surface: "#1F1D2E", Panel: "#26233A", Foreground: "#E0DEF4", Primary: "#C4A7E7", Warning: "#F5C177", Error: "#EA6F92", Success: "#9CCFD8", Accent: "#EBBCBA", Muted: "#A3A2A7", FooterBackground: "#26233A", FooterForeground: "#E0DEF4", FooterKeyForeground: "#EBBCBA", FooterDescriptionForeground: "#E0DEF4"},
	{Name: "rose-pine-moon", Dark: true, Background: "#232136", Surface: "#2A273F", Panel: "#393552", Foreground: "#E0DEF4", Primary: "#C4A7E7", Warning: "#F5C177", Error: "#EA6F92", Success: "#9CCFD8", Accent: "#EA9A97", Muted: "#A7A6AF", FooterBackground: "#393552", FooterForeground: "#E0DEF4", FooterKeyForeground: "#EA9A97", FooterDescriptionForeground: "#E0DEF4"},
	{Name: "rose-pine-dawn", Dark: false, Background: "#FAF4ED", Surface: "#FFFAF3", Panel: "#F2E9E1", Foreground: "#575279", Primary: "#907AA9", Warning: "#E99D34", Error: "#B4637A", Success: "#56949F", Accent: "#D6827E", Muted: "#64625F", FooterBackground: "#F2E9E1", FooterForeground: "#575279", FooterKeyForeground: "#D7827E", FooterDescriptionForeground: "#575279"},
	{Name: "atom-one-dark", Dark: true, Background: "#282C34", Surface: "#3B414D", Panel: "#4F5666", Foreground: "#ABB2BF", Primary: "#61AFEF", Warning: "#DDB25B", Error: "#EF6262", Success: "#62F062", Accent: "#A378C2", Muted: "#A9ABAE", FooterBackground: "#4F5666", FooterForeground: "#ABB2BF", FooterKeyForeground: "#A378C2", FooterDescriptionForeground: "#ABB2BF"},
	{Name: "atom-one-light", Dark: false, Background: "#FAFAFA", Surface: "#E0E0E0", Panel: "#CCCCCC", Foreground: "#383A42", Primary: "#4078F2", Warning: "#D7D938", Error: "#F13F3F", Success: "#6BF23F", Accent: "#BE9232", Muted: "#646464", FooterBackground: "#CCCCCC", FooterForeground: "#383A42", FooterKeyForeground: "#BF9232", FooterDescriptionForeground: "#383A42"},
	{Name: "ansi-dark", Dark: true, Background: "", Surface: "", Panel: "", Foreground: "", Primary: "4", Warning: "3", Error: "1", Success: "2", Accent: "2", Muted: "8", FooterBackground: "", FooterForeground: "", FooterKeyForeground: "5", FooterDescriptionForeground: ""},
	{Name: "ansi-light", Dark: false, Background: "", Surface: "", Panel: "", Foreground: "", Primary: "4", Warning: "9", Error: "1", Success: "2", Accent: "5", Muted: "7", FooterBackground: "", FooterForeground: "", FooterKeyForeground: "5", FooterDescriptionForeground: ""},
}

// activeTheme defaults to ansi-dark before settings are loaded (matching
// tui.py's on_mount hard-coding `self.theme = "ansi-dark"`), but unlike the
// Python original a picked theme *is* persisted to settings.json
// (Settings.Theme) and restored on the next launch via setActiveTheme in
// main.go - a deliberate deviation, since forgetting the user's theme pick
// on every restart reads as a bug rather than a faithful behavior to
// preserve. (themes[0] is "textual-dark", not ansi-dark - Textual's own
// registration order, kept as-is since that's the order the Theme command
// palette page lists them in too - so this can't just be themes[0].)
var activeTheme = defaultTheme()

func defaultTheme() Theme {
	t, ok := themeByName("ansi-dark")
	if !ok {
		return themes[0]
	}
	return t
}

func themeByName(name string) (Theme, bool) {
	for _, t := range themes {
		if t.Name == name {
			return t, true
		}
	}
	return Theme{}, false
}

// setActiveTheme switches the active theme and rebuilds every derived style
// (see applyTheme in style.go), returning false (no-op) if the name isn't
// recognized.
func setActiveTheme(name string) bool {
	t, ok := themeByName(name)
	if !ok {
		return false
	}
	activeTheme = t
	applyTheme(t)
	return true
}
