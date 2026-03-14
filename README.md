portyanka
=========

[![CI](https://github.com/odysa/portyanka/actions/workflows/ci.yml/badge.svg)](https://github.com/odysa/portyanka/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/odysa/portyanka)](https://goreportcard.com/report/github.com/odysa/portyanka)
[![License](https://img.shields.io/github/license/odysa/portyanka)](https://github.com/odysa/portyanka/blob/main/LICENSE)

A minimal Go TUI for inspecting listening TCP ports and killing processes. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**Enhanced features:** Sorting, filtering, expandable details with connection counts, and a beautiful TUI powered by Bubble Tea.

![demo](static/demo.gif)

## Features

- List all listening TCP ports with PID, process name, protocol, address, and port
- **Sorting** by PID, port, or process name (press `p`, `n`, `o`)
- **Expandable details** showing full command line and connection counts (press `e`)
- Real-time filtering by process name or port number (case-insensitive)
- Kill processes with SIGTERM or SIGKILL with confirmation
- Vim-style navigation (j/k or arrow keys)
- Wrapping selection and scrolling for long lists
- Beautiful TUI with Bubble Tea framework
- Clean terminal restoration on exit

## Installation

### Homebrew
```bash
brew install odysa/tap/portyanka
```

### Using install script
```bash
curl -fsSL https://raw.githubusercontent.com/odysa/portyanka/main/install.sh | sh
```

### Using go install
```bash
go install github.com/odysa/portyanka@latest
```

### Build from source
```bash
git clone https://github.com/odysa/portyanka.git
cd portyanka
go build
```

**Requirements:** Go 1.23+ for building from source

## Usage

### Keybindings

#### Normal Mode
| Key | Action |
|-----|--------|
| `j` / `k` or arrow keys | Move selection up/down |
| `e` | Expand/collapse details for selected row |
| `Enter` | Open action menu |
| `/` | Filter by name or port |
| `p` | Sort by PID (toggle ascending/descending) |
| `n` | Sort by process name (toggle ascending/descending) |
| `o` | Sort by port (toggle ascending/descending) |
| `K` | Kill selected process (SIGTERM) |
| `F` | Force kill selected process (SIGKILL) |
| `r` | Refresh port list |
| `q` / `Esc` | Quit |

#### Filter Mode
| Key | Action |
|-----|--------|
| `Enter` | Apply filter |
| `Esc` | Cancel filter |

#### Action Menu & Dialogs
| Key | Action |
|-----|--------|
| `j` / `k` or arrow keys | Navigate options |
| `Enter` | Confirm selection |
| `Esc` / `q` | Close/Cancel |
| `y` / `n` | Confirm/Cancel kill (in confirmation dialog) |

## Requirements

- macOS or Linux (x86_64 or aarch64)
- `lsof` and `kill` in PATH

## License

MIT. See [LICENSE](LICENSE) for details.
