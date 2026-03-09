---
name: ttg
description: >
  Terminal CLI for TickTick — add tasks, list projects, mark things done, and manage your inbox without leaving the shell. Built in Go, uses the official TickTick Open API (OAuth2). Binary is `ttg`.
version: 1.0.0
author: dhruvkelawala
homepage: https://github.com/dhruvkelawala/tt
metadata:
  openclaw:
    emoji: ✅
    requires:
      bins:
        - ttg
    install:
      - id: source
        kind: shell
        label: "Build and install ttg from source"
        run: |
          git clone https://github.com/dhruvkelawala/tt /tmp/ttg-install
          cd /tmp/ttg-install
          make install
          rm -rf /tmp/ttg-install
---

# ttg — TickTick CLI

Use `ttg` to manage TickTick tasks and projects from the terminal.

> **Binary name is `ttg`** — not `tt`. Always invoke as `ttg`.

## Prerequisites

1. Go 1.21+ installed (`brew install go` on macOS)
2. `~/.local/bin` on your `$PATH`
3. TickTick developer app credentials — register at [developer.ticktick.com](https://developer.ticktick.com/manage)

## Setup (one-time)

Create `~/.config/ttg/config.json`:

```json
{
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "timezone": "Europe/London"
}
```

Then authenticate:

```bash
ttg auth login
```

Token is stored at `~/.config/ttg/token.json`.

## Trigger phrases

Use this skill when the user asks to:

- "add a task", "create a task", "remind me to..."
- "list my tasks", "what's in my inbox", "show tasks due today"
- "mark task done", "complete a task", "check off..."
- "delete a task", "remove a task"
- "list my projects", "show projects"
- "edit a task", "change priority", "set due date"

## Commands

### Task management

```bash
ttg task list                              # Inbox tasks
ttg task list --all                        # All projects
ttg task list --project "Work"             # Filter by project
ttg task list --due today                  # Due today
ttg task list --due tomorrow               # Due tomorrow
ttg task list --priority high              # High priority only
ttg task list --json                       # JSON output

ttg task add "Task title"
ttg task add "Task title" --project "Work" --priority high --due "tomorrow 9am"
ttg task add "Task title" --due "next monday" --priority medium

ttg task get <task-id>                     # Get task details
ttg task done <task-id>                    # Mark complete
ttg task delete <task-id>                  # Delete task
ttg task edit <task-id> --title "New title"
ttg task edit <task-id> --priority medium --due "2026-04-01"
```

### Project management

```bash
ttg project list                           # List all projects
ttg project get <project-id>              # Project details
```

### Auth

```bash
ttg auth login                             # OAuth2 browser flow
ttg auth status                            # Check token validity
```

## Due date formats

| Input | Meaning |
|-------|---------|
| `today`, `tomorrow` | Midnight of that day |
| `next monday` | Following Monday |
| `3pm`, `tomorrow 3pm` | Specific time |
| `in 2 days`, `in 3 hours` | Relative offset |
| `2026-03-20` | ISO date |
| `2026-03-20T15:00:00` | ISO datetime |

## Priority values

`none` (default) · `low` · `medium` · `high`

## JSON output

Any command supports `--json` / `-j` for scripting:

```bash
ttg task list --json | jq '.[].title'
ttg task list --all --json | jq '.[] | select(.priority == 5)'
```

## Agent instructions

1. For "add task" requests: extract title, project (if mentioned), priority (if mentioned), due date (if mentioned). Use natural language for due date — `ttg` parses it.
2. For "list tasks": default to inbox unless user specifies a project or `--all`.
3. For "mark done": you'll need the task ID. Run `ttg task list --json` first to find it, then `ttg task done <id>`.
4. For "delete task": same — find ID via list, then delete. Confirm with user before deleting.
5. Always use `--json` when you need to parse output programmatically.
