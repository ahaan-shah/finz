# tally

A terminal personal finance tracker — a Go/[bubbletea](https://github.com/charmbracelet/bubbletea) port of a Python/[Textual](https://textual.textualize.io/) app of the same name, built for one-binary distribution.

## What it does

- Tracks expenses and income (one list, sign tells them apart)
- A year → month → ledger drill-down, with per-period totals at every step
- A monthly budget gauge and category/month spending charts
- 21 built-in color themes, switchable from the command palette
- Export to CSV, XLSX, or JSON — always dropped in `~/Downloads`

Everything is stored locally — no account, no server, no sync.

## Installation

You'll need [Go](https://go.dev/dl/) installed (1.21+).

```bash
go install github.com/ahaan-shah/tally@latest
```

Make sure `$(go env GOPATH)/bin` (usually `~/go/bin`) is on your `PATH`, then run:

```bash
tally
```

### Building from source

```bash
git clone https://github.com/ahaan-shah/tally.git
cd tally
go build -o tally .
./tally
```

## Quick start

First launch seeds a few sample transactions so the ledger isn't empty. From there:

| Key | Does what |
|---|---|
| `↑` / `↓` | move the highlight (year/month picker) or the ledger cursor |
| `enter` | select (year/month picker) |
| `a` / `e` / `d` | add / edit / delete a transaction (ledger stage) |
| `tab` | switch focus to/from the budget input (ledger stage) |
| `alt+←` / `alt+→` | step back / forward a stage |
| `ctrl+p` | command palette — currency, themes, export, keybindings |
| `q` / `ctrl+c` | quit |

Your data lives at `~/.config/tally/` (or your OS's equivalent). Exports (CSV/XLSX/JSON, via the command palette's "Export" entry) always land in `~/Downloads`, dated like `export-2026-07-14.csv`.

## Relation to the Python original

This is a faithful port, not a redesign — navigation, layout, colors, and keybinds all aim to match the original one-to-one, down to reproducing Textual's exact border glyphs and widget layout math. Two intentional differences: where data is stored (a standalone Go binary has no fixed "next to the script" location the way a cloned Python repo does, so `tally` uses `~/.config/tally` instead), and the add/edit/delete modals plus the command palette, which are styled to match [splitsy](https://github.com/ahaan-shah/splitsy) — tally's Go sibling — instead of Textual's own modal chrome.
