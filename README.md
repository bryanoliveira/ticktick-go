# tt - TickTick CLI

A Go CLI binary for managing TickTick tasks from the terminal.

## Installation

```bash
make install
```

This installs `tt` to `~/.local/bin/tt`.

## Usage

### Authentication

```bash
tt auth login        # OAuth2 browser flow
tt auth status       # Show auth status
tt auth logout       # Delete stored token
```

### Tasks

```bash
# List tasks (inbox by default)
tt task list
tt task list --project "Work"
tt task list --all
tt task list --due today
tt task list --priority high

# Add task
tt task add "Buy milk"
tt task add "Deploy app" --project "Work" --priority high --due "tomorrow 3pm"

# Task details
tt task get <id>

# Complete/delete
tt task done <id>
tt task delete <id>

# Edit task
tt task edit <id> --title "New title" --priority medium
```

### Projects

```bash
tt project list
tt project get <id>
```

### Quick Add

```bash
tt "Buy milk" --project Work --priority high --due tomorrow
```

### JSON Output

Add `--json` or `-j` flag to any command for JSON output.

## Configuration

Config is stored at `~/.config/tt/config.json`:

```json
{
  "timezone": "Europe/London",
  "default_project": "inbox",
  "client_id": "24qE700R7e12YnSNWj",
  "client_secret": "4kF89Zm77tWhMvhNq0TiL4PTxavRTdCJ"
}
```

Token is stored at `~/.config/tt/token.json`.

## Date Parsing

Supported date formats:
- `today`, `tomorrow`, `yesterday`
- `next monday`, `next friday`
- `3pm`, `tomorrow 3pm`, `9am`
- ISO: `2026-03-20`, `2026-03-20T15:00:00`
- Relative: `in 2 days`, `in 3 hours`
