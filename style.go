package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// All of the vars below are deliberately mutable (not const, not built
// once) - applyTheme rebuilds every one of them in place whenever the
// active theme changes (Textual re-colors every themed widget the instant
// you pick a new theme from the command palette; this is how the Go port
// gets the same behavior without a CSS engine).
var (
	colorPrimary lipgloss.Color
	colorWarning lipgloss.Color
	colorError   lipgloss.Color
	colorSuccess lipgloss.Color
	colorAccent  lipgloss.Color
	colorMuted   lipgloss.Color

	styleBold    lipgloss.Style
	styleMuted   lipgloss.Style
	styleError   lipgloss.Style
	styleSuccess lipgloss.Style
	styleWarning lipgloss.Style
	styleAccent  lipgloss.Style
	stylePrimary lipgloss.Style

	// styleHeader/styleFooter are the top/bottom chrome bars - background
	// $panel (footer separately themeable, see Theme.FooterBackground -
	// solarized is the one built-in theme that actually overrides it).
	styleHeader            lipgloss.Style
	styleHeaderSubtitle    lipgloss.Style
	styleFooter            lipgloss.Style
	styleFooterKey         lipgloss.Style
	styleFooterDescription lipgloss.Style
	styleFooterSeparator   lipgloss.Style

	// Modal chrome (add/edit/delete transaction, command palette) is
	// deliberately styled to match splitsy - tally's Go sibling - one for
	// one, not Textual's own modal look: lipgloss's stock ThickBorder()
	// preset rather than Textual's blocky "thick" glyphs, plain label+
	// value field rows with no per-field box, and plain-text buttons.
	styleModalBox       lipgloss.Style
	styleModalBoxDanger lipgloss.Style
	styleModalTitle     lipgloss.Style

	// styleModalFilterBox boxes the command palette's search Input -
	// splitsy's rounded-border convention, distinct from the tall-bordered
	// boxes the year/month picker still uses (that picker stays faithful
	// to Textual's own Input/OptionList look, which nothing here asked to
	// change).
	styleModalFilterBox lipgloss.Style

	// styleFieldLabel/Focused are the add/edit form's row labels - fixed
	// width, muted normally, bold+accent (with a "▸" prefix) while that
	// field has focus - matching splitsy's renderFieldRow exactly.
	styleFieldLabel   lipgloss.Style
	styleFieldFocused lipgloss.Style

	// Plain-text buttons, no box - matching splitsy's stripped-down modal
	// buttons (which themselves mirror Textual's Button CSS reset).
	styleButtonCancel lipgloss.Style
	styleButtonSave   lipgloss.Style
	styleButtonDelete lipgloss.Style

	// styleTableCursor is the DataTable's focused-cursor row highlight
	// ($block-cursor-background, which equals $primary for every built-in
	// theme except rose-pine-dawn - approximated as $primary throughout).
	styleTableCursor lipgloss.Style

	appBackground lipgloss.Color
	appForeground lipgloss.Color

	// canvasRepaint/panelRepaint/footerRepaint/surfaceRepaint/
	// primaryRepaint are each a raw "\x1b[48;2;...m\x1b[38;2;...m" pair for
	// one region's background+foreground (general canvas, header/footer
	// bar, ledger table's selected row) - every lipgloss Render() call
	// closes with a full fg+bg reset regardless of what it actually set,
	// so a styled fragment that isn't the last thing on its line leaves
	// whatever follows it sitting on the terminal's own unpainted
	// background instead of the intended one. repaintWith splices the
	// right region's colors back in right after every reset it finds, and
	// - since later repaintWith calls only re-touch the *original* reset
	// bytes, never anything a previous call appended - repainting a small
	// region first and the whole frame with the general canvas colors
	// last (see app.go's View()) naturally leaves the small region's own
	// colors as the ones that actually win, without needing to protect it
	// from the later, broader pass.
	canvasRepaint  string
	panelRepaint   string
	footerRepaint  string
	surfaceRepaint string
	primaryRepaint string

	// primaryMarker is primaryRepaint's background code alone - the exact
	// prefix bubbles/table's styles.Selected wrap emits for the cursor
	// row, used to tell that row's already-composed line apart from every
	// other line when repainting the table's rendered output (see
	// repaintTableView).
	primaryMarker string
)

// tallBorder reproduces Textual's built-in "tall" border character set
// exactly (textual/_border.py: BORDER_LOCATIONS["tall"] /
// BORDER_CHARS["tall"]) - the default border Input and OptionList both
// draw, distinct from any of lipgloss's own preset borders.
var tallBorder = lipgloss.Border{
	Top:         "▔",
	Bottom:      "▁",
	Left:        "▊",
	Right:       "▎",
	TopLeft:     "▊",
	TopRight:    "▎",
	BottomLeft:  "▊",
	BottomRight: "▎",
}

func init() {
	applyTheme(activeTheme)
}

// applyTheme rebuilds every color/style package var from t. Called once at
// startup and again each time the user picks a new theme from the command
// palette (Textual's own built-in "Theme" system command, in this port's
// case just another page of the ctrl+p palette).
func applyTheme(t Theme) {
	colorPrimary = t.Primary
	colorWarning = t.Warning
	colorError = t.Error
	colorSuccess = t.Success
	colorAccent = t.Accent
	colorMuted = t.Muted
	appBackground = t.Background
	appForeground = t.Foreground

	// bg conditionally applies the canvas background to a style so its own
	// closing reset (every lipgloss Render() call resets fg *and* bg, even
	// when only one was ever set) can't leave a gap of the terminal's own
	// background showing through mid-line. No-op for the ansi-* themes,
	// which deliberately leave t.Background empty.
	bg := func(s lipgloss.Style) lipgloss.Style {
		if t.Background == "" {
			return s
		}
		return s.Background(t.Background)
	}

	styleBold = bg(lipgloss.NewStyle().Bold(true))
	styleMuted = bg(lipgloss.NewStyle().Foreground(colorMuted))
	styleError = bg(lipgloss.NewStyle().Foreground(colorError))
	styleSuccess = bg(lipgloss.NewStyle().Foreground(colorSuccess))
	styleWarning = bg(lipgloss.NewStyle().Foreground(colorWarning))
	styleAccent = bg(lipgloss.NewStyle().Foreground(colorAccent))
	stylePrimary = bg(lipgloss.NewStyle().Foreground(colorPrimary))

	headerBG := t.Panel
	styleHeader = lipgloss.NewStyle().Foreground(t.Foreground)
	if headerBG != "" {
		styleHeader = styleHeader.Background(headerBG)
	}
	styleHeaderSubtitle = styleHeader.Faint(true)

	footerBG := t.FooterBackground
	styleFooter = lipgloss.NewStyle().Foreground(t.FooterForeground)
	if footerBG != "" {
		styleFooter = styleFooter.Background(footerBG)
	}
	styleFooterKey = styleFooter.Bold(true).Foreground(t.FooterKeyForeground)
	styleFooterDescription = styleFooter.Foreground(t.FooterDescriptionForeground)
	styleFooterSeparator = styleFooter.Foreground(t.Foreground).Faint(true)

	styleModalBox = bg(lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(colorAccent).Padding(1, 3))
	styleModalBoxDanger = bg(lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(colorError).Padding(1, 3))
	styleModalTitle = bg(lipgloss.NewStyle().Bold(true).MarginBottom(1))
	styleModalFilterBox = bg(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).Padding(0, 1))

	styleFieldLabel = bg(lipgloss.NewStyle().Width(10).Foreground(colorMuted))
	styleFieldFocused = bg(lipgloss.NewStyle().Width(10).Bold(true).Foreground(colorAccent))
	styleButtonCancel = bg(lipgloss.NewStyle().Foreground(colorMuted))
	styleButtonSave = bg(lipgloss.NewStyle().Bold(true).Foreground(colorPrimary))
	styleButtonDelete = bg(lipgloss.NewStyle().Bold(true).Foreground(colorError))

	primaryFG := cursorForeground(t)
	styleTableCursor = lipgloss.NewStyle().Bold(true).Background(colorPrimary).Foreground(primaryFG)

	canvasRepaint = ""
	panelRepaint = ""
	footerRepaint = ""
	surfaceRepaint = ""
	primaryRepaint = ""
	primaryMarker = ""
	if t.Background != "" {
		canvasRepaint = ansiTrueColor(48, t.Background) + ansiTrueColor(38, t.Foreground)
	}
	if t.Panel != "" {
		panelRepaint = ansiTrueColor(48, t.Panel) + ansiTrueColor(38, t.Foreground)
	}
	if t.FooterBackground != "" {
		footerRepaint = ansiTrueColor(48, t.FooterBackground) + ansiTrueColor(38, t.FooterForeground)
	}
	if t.Surface != "" {
		surfaceRepaint = ansiTrueColor(48, t.Surface) + ansiTrueColor(38, t.Foreground)
		primaryMarker = ansiTrueColor(48, colorPrimary)
		primaryRepaint = ansiTrueColor(48, colorPrimary) + ansiTrueColor(38, primaryFG)
	}
}

// cursorForeground approximates Textual's `auto 87%` block-cursor
// foreground (auto-selected black/white for contrast, so close to solid
// black/white at 87% opacity it's indistinguishable in a terminal) by
// picking whichever of black/white contrasts more with the cursor's own
// background - for the ansi-* themes there's no fixed hex to test, so it
// falls back to the theme's Dark flag instead.
func cursorForeground(t Theme) lipgloss.Color {
	if !strings.HasPrefix(string(t.Primary), "#") {
		if t.Dark {
			return lipgloss.Color("0")
		}
		return lipgloss.Color("15")
	}
	if relativeLuminance(string(t.Primary)) > 0.5 {
		return lipgloss.Color("0")
	}
	return lipgloss.Color("15")
}

func relativeLuminance(hex string) float64 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	// Standard perceptual luminance weighting.
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255
}

// ansiTrueColor builds a raw truecolor SGR sequence (kind 38 for
// foreground, 48 for background) for a "#RRGGBB" lipgloss.Color, bypassing
// lipgloss/termenv entirely so repaintCanvas can splice it directly into
// an already-rendered string.
func ansiTrueColor(kind int, c lipgloss.Color) string {
	hex := strings.TrimPrefix(string(c), "#")
	if len(hex) != 6 {
		return ""
	}
	r, err1 := strconv.ParseUint(hex[0:2], 16, 8)
	g, err2 := strconv.ParseUint(hex[2:4], 16, 8)
	b, err3 := strconv.ParseUint(hex[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return ""
	}
	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", kind, r, g, b)
}

// repaintWith re-stamps code (one of the *Repaint vars above) right after
// every reset embedded in s - see the canvasRepaint doc comment for why
// this is necessary. A no-op if code is empty (the ansi-* themes never
// populate any of these).
func repaintWith(s, code string) string {
	if code == "" {
		return s
	}
	return strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m"+code)
}

// repaintCanvas is repaintWith for the general app-canvas background -
// used once, on the fully-composed frame, in app.go's View().
func repaintCanvas(s string) string {
	return repaintWith(s, canvasRepaint)
}

// repaintLedgerView fixes the ledger table's own version of the
// reset-leak: each row is built from several independently-colored cells
// (date/category plain, amount/balance colored), and the selected row
// wraps that whole already-composed line in one more Render() call, so
// everything after the first colored cell goes unpainted the instant a
// row is selected (every lipgloss Render() call resets fg+bg regardless
// of what it actually set). Since each cell's own prefix otherwise
// repaints its own background correctly, the only way to tell rows apart
// after the fact is to look for the exact background prefix
// styleTableCursor itself would have emitted (primaryMarker) - present
// only on the selected row's line. Row 0 is always the header (Panel
// background, not Surface).
func repaintLedgerView(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		switch {
		case i == 0:
			lines[i] = repaintWith(line, panelRepaint)
		case primaryMarker != "" && strings.Contains(line, primaryMarker):
			lines[i] = repaintWith(line, primaryRepaint)
		default:
			lines[i] = repaintWith(line, surfaceRepaint)
		}
	}
	return strings.Join(lines, "\n")
}

// barStyle/mutedBarStyle carry the app's canvas background (see bg in
// applyTheme) so a chart bar segment's own closing reset doesn't leave the
// track/label text after it sitting on the terminal's own background.
func barStyle(c lipgloss.Color) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(c)
	if appBackground != "" {
		s = s.Background(appBackground)
	}
	return s
}

func mutedBarStyle() lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(colorMuted).Faint(true)
	if appBackground != "" {
		s = s.Background(appBackground)
	}
	return s
}
