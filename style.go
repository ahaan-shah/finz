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

	// styleFooter is the bottom chrome bar - background $panel by default
	// (see Theme.FooterBackground - solarized is the one built-in theme
	// that actually overrides it).
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

	// styleFilterBox boxes every boxed single-line Input in the app (the
	// year/month picker's filter, the budget input, the command palette's
	// search box) - splitsy's rounded-border convention exactly, replacing
	// Textual's much heavier "tall" block-glyph border this used to draw.
	styleFilterBox lipgloss.Style

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

	// styleLedgerHeaderText/styleLedgerError/styleLedgerSuccess are plain
	// foreground-only styles with *no* baked-in background, unlike the
	// package-level styleBold/styleError/styleSuccess (which bake in the
	// general canvas background via applyTheme's bg helper) - every
	// ledger row's background comes entirely from repaintLedgerView
	// (panel for the header, primary for the selected row, surface for
	// everything else), and a cell that sets its own background fights
	// that: it was overriding the selected row's highlight for whichever
	// cell happened to be colored (the amount/balance columns), leaving a
	// canvas-colored gap in the middle of an otherwise fully-highlighted
	// row.
	styleLedgerHeaderText lipgloss.Style
	styleLedgerError      lipgloss.Style
	styleLedgerSuccess    lipgloss.Style

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
)

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
	styleFilterBox = bg(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).Padding(0, 1))

	styleFieldLabel = bg(lipgloss.NewStyle().Width(10).Foreground(colorMuted))
	styleFieldFocused = bg(lipgloss.NewStyle().Width(10).Bold(true).Foreground(colorAccent))
	styleButtonCancel = bg(lipgloss.NewStyle().Foreground(colorMuted))
	styleButtonSave = bg(lipgloss.NewStyle().Bold(true).Foreground(colorPrimary))
	styleButtonDelete = bg(lipgloss.NewStyle().Bold(true).Foreground(colorError))

	primaryFG := cursorForeground(t)
	styleTableCursor = lipgloss.NewStyle().Bold(true).Background(colorPrimary).Foreground(primaryFG)
	styleLedgerHeaderText = lipgloss.NewStyle().Bold(true)
	styleLedgerError = lipgloss.NewStyle().Foreground(colorError)
	styleLedgerSuccess = lipgloss.NewStyle().Foreground(colorSuccess)

	canvasRepaint = ""
	panelRepaint = ""
	footerRepaint = ""
	surfaceRepaint = ""
	primaryRepaint = ""
	if t.Background != "" {
		canvasRepaint = ansiTrueColor(48, t.Background) + ansiTrueColor(38, t.Foreground)
	}
	if t.Panel != "" {
		panelRepaint = ansiTrueColor(48, t.Panel) + ansiTrueColor(38, t.Foreground)
	}
	if t.FooterBackground != "" {
		footerRepaint = ansiTrueColor(48, t.FooterBackground) + ansiTrueColor(38, t.FooterForeground)
	}
	// Computed unconditionally (not gated behind t.Surface, unlike the
	// other *Repaint vars above): the cursor row's highlight needs to stay
	// visible even under the ansi-* themes, which otherwise force no
	// background at all.
	primaryRepaint = ansiColorCode(48, colorPrimary) + ansiColorCode(38, primaryFG)
	if t.Surface != "" {
		surfaceRepaint = ansiTrueColor(48, t.Surface) + ansiTrueColor(38, t.Foreground)
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

// sgrParamTrueColor returns just the numeric SGR parameter (no leading
// "\x1b[" or trailing "m") for a truecolor "#RRGGBB" lipgloss.Color - kind
// 38 for foreground, 48 for background. Kept separate from
// ansiTrueColor/ansiColorCode below only so they can share this logic.
func sgrParamTrueColor(kind int, c lipgloss.Color) string {
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
	return fmt.Sprintf("%d;2;%d;%d;%d", kind, r, g, b)
}

// sgrParam is sgrParamTrueColor generalized to also handle the ansi-*
// themes' plain decimal palette-index colors ("0".."15", standard/bright
// ANSI), which have no hex to convert - used for the ledger's cursor-row
// repaint (primaryRepaint), which (like Textual's own block-cursor
// highlight) needs to stay visible even in the themes that otherwise
// leave every other background untouched, since "which row is selected"
// has to be visible regardless of theme.
func sgrParam(kind int, c lipgloss.Color) string {
	if p := sgrParamTrueColor(kind, c); p != "" {
		return p
	}
	n, err := strconv.Atoi(string(c))
	if err != nil || n < 0 || n > 15 {
		return ""
	}
	base := 30
	if kind == 48 {
		base = 40
	}
	if n >= 8 {
		base += 60 // bright range: 90-97 / 100-107
		n -= 8
	}
	return strconv.Itoa(base + n)
}

// ansiTrueColor/ansiColorCode wrap sgrParamTrueColor/sgrParam into a
// complete escape sequence, for splicing directly into an already-rendered
// string (repaintWith).
func ansiTrueColor(kind int, c lipgloss.Color) string {
	p := sgrParamTrueColor(kind, c)
	if p == "" {
		return ""
	}
	return "\x1b[" + p + "m"
}

func ansiColorCode(kind int, c lipgloss.Color) string {
	p := sgrParam(kind, c)
	if p == "" {
		return ""
	}
	return "\x1b[" + p + "m"
}

// repaintWith re-stamps code (one of the *Repaint vars above) at the start
// of every line and right after every reset embedded in it - see the
// canvasRepaint doc comment for why this is necessary. A no-op if code is
// empty (the ansi-* themes never populate any of these).
//
// The leading stamp matters for lines built by centering/padding an
// already-rendered ANSI fragment (lipgloss.Style.Align(Center), say): the
// padding spaces lipgloss inserts *before* the fragment's first escape
// code carry no color of their own, so without an explicit prefix they'd
// sit on whatever the terminal's ambient background happens to be instead
// of the theme's - this is what left a plain-black block to the left of a
// centered header title on every non-ansi-* theme.
//
// Operates per line (s may be a single line or a whole multi-line frame -
// repaintCanvas calls it on the entire composed frame in one shot) and
// deliberately leaves each *line's own trailing* reset alone: SGR
// background state isn't cleared by a bare "\n" in a real terminal, so a
// line that ends in "reset, then immediately re-set the background again"
// leaves that color active for whatever comes next - a blank spacer line,
// or the next widget's own first character, which inherits it if that
// character's style never set an explicit background of its own (border
// glyphs in particular: BorderForeground alone doesn't imply a matching
// BorderBackground). Only re-stamping *interior* resets (an amount cell
// finishing partway through a ledger row, say) and leaving the line's
// final reset as a clean reset is what actually stops that bleed.
func repaintWith(s, code string) string {
	if code == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		out := code + strings.ReplaceAll(line, "\x1b[0m", "\x1b[0m"+code)
		out = strings.TrimSuffix(out, code)
		lines[i] = out
	}
	return strings.Join(lines, "\n")
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
// of what it actually set). cursorLine is which line (already known by
// the caller - lines[0] is always the header) gets the cursor's own
// primary-background repaint instead of the ordinary surface one.
//
// This used to detect the cursor line by searching each line's text for
// the literal background escape code styleTableCursor would have
// emitted - that broke silently on some themes, because lipgloss's own
// truecolor rendering (which goes through a color-space round trip) can
// round a hex value to a *different* decimal RGB than a direct hex-to-int
// parse of the same "#RRGGBB" string (off by 1 in one channel), so the
// hand-built marker byte-for-byte never matched what lipgloss actually
// wrote. Tracking the row by index instead of by sniffing its rendered
// color sidesteps that whole class of bug.
func repaintLedgerView(s string, cursorLine int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		switch {
		case i == 0:
			lines[i] = repaintWith(line, panelRepaint)
		case i == cursorLine:
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
