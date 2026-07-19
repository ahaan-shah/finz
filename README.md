# Pear

**A terminal personal finance tracker.**

Pretty much entirely vibecoded 👾

## Preview ✨

<video src="https://github.com/user-attachments/assets/a1d7adc8-ad2a-4db0-9490-a167a106afa4" controls></video>

## Features

- Tracks income and expenses
- Built in currency selector menu
- A year → month → ledger drill-down
- A monthly budget setting and category/month spending charts
- 21 built-in color themes, switchable from the command palette
- Export to CSV, XLSX, or JSON — always dropped in `~/Downloads`

Everything is stored locally — no account, no hassle.

## Prerequisites

You'll need [Go](https://go.dev/dl/) installed (1.21+).

That's it.

### Don't have Go yet?

#### Arch
```bash
sudo pacman -S go
```

#### Debian / Ubuntu (and friends)
```
sudo apt install golang-go
```

#### macOS (Homebrew)
```
brew install go
```

Works the same on:

- Linux
- MacOS
- Windows

## Installation

```bash
go install github.com/ahaan-shah/pear@latest
```

This drops a `pear` binary in `$(go env GOPATH)/bin` (usually `~/go/bin`). Make sure that's on your `PATH`:

```bash
# add to your shell config (~/.zshrc, ~/.bashrc, whatever you're rocking)
export PATH="$HOME/go/bin:$PATH"
```

Reload your shell, then just run:

```bash
pear
```

### Building from source (if you wanna clone this repo)

```bash
git clone https://github.com/ahaan-shah/pear.git
cd pear
go build -o pear .
./pear
```

## Quick start

First launch seeds a few sample transactions so the ledger isn't empty. From there:

| Key | Does what |
|---|---|
| `↑` / `↓` | scroll up or down |
| `enter` | select choice |
| `a` / `e` / `d` | add / edit / delete a transaction (ledger stage) |
| `tab` | switch focus to/from the budget input (ledger stage) |
| `alt+←` / `alt+→` | step back / forward a stage |
| `ctrl+p` | command palette — currency, themes, export, keybindings |
| `q` / `ctrl+c` | quit |

Your data lives at `~/.config/pear/` (or your OS's equivalent). Exports always land in `~/Downloads`.

## License

[MIT](LICENSE) — Do what you want with it.
