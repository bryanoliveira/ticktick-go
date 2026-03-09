<div align="center">

# ttg

**A fast, minimal CLI for [TickTick](https://ticktick.com) — built in Go.**

Add tasks, manage projects, and check off your day without leaving the terminal.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)

</div>

---

## Features

- ✅ Full task CRUD — list, add, edit, complete, delete
- 📁 Project management
- 🗓 Natural language due dates (`tomorrow 3pm`, `next monday`, `in 2 days`)
- 🔺 Priority support (`low`, `medium`, `high`)
- 📤 JSON output for scripting
- 🔐 OAuth2 auth via TickTick Open API

---

## Installation

**From source (macOS, Linux, any platform):**

```bash
git clone https://github.com/dhruvkelawala/tt
cd tt
make install   # builds and copies to ~/.local/bin/ttg
```

Make sure `~/.local/bin` is on your `$PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

> **Platform note:** `make install` compiles from source and works on macOS (Intel + Apple Silicon), Linux (amd64, arm64), and any Go-supported platform. Requires Go 1.21+.

---

## Setup

1. Register an app at [developer.ticktick.com](https://developer.ticktick.com/manage) to get your **Client ID** and **Client Secret**.

2. Create `~/.config/ttg/config.json`:

```json
{
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "timezone": "Europe/London"
}
```

3. Authenticate:

```bash
ttg auth login
```

This opens a browser for OAuth2 and stores your token at `~/.config/ttg/token.json`.

---

## Usage

### Tasks

```bash
ttg task list                          # Inbox (default)
ttg task list --all                    # All tasks
ttg task list --project "Work"         # By project
ttg task list --due today              # Due today
ttg task list --priority high          # By priority
ttg task list --json                   # JSON output

ttg task add "Buy milk"
ttg task add "Ship feature" --project "Work" --priority high --due "tomorrow 9am"

ttg task get <id>                      # Task details
ttg task done <id>                     # Mark complete
ttg task delete <id>                   # Delete
ttg task edit <id> --title "New title" --priority medium
```

### Projects

```bash
ttg project list
ttg project get <id>
```

### JSON / scripting

Any command accepts `--json` / `-j`:

```bash
ttg task list --json | jq '.[].title'
```

---

## Due Date Formats

| Input | Meaning |
|-------|---------|
| `today`, `tomorrow` | Midnight of that day |
| `next monday` | Following Monday |
| `3pm`, `tomorrow 3pm` | Specific time |
| `in 2 days`, `in 3 hours` | Relative offset |
| `2026-03-20` | ISO date |
| `2026-03-20T15:00:00` | ISO datetime |

---

## Priority

`none` (default) · `low` · `medium` · `high`

---

## License

MIT
