# Executive System

## Purpose

The executive assistant manages Ashley's life: planning, task tracking,
calendar, priorities, decisions, and personal knowledge.

## Model

A single, long-lived OpenCode environment. Type `assistant`, OpenCode
opens in `~/Executive/`, and you manage your life. There is one
environment, not many.

## Working directory

`~/Executive/`

## Directory structure

```
~/Executive/
├── opencode.json              # OpenCode config (provider, model, MCP servers)
├── AGENTS.md                  # System prompt / behavioral rules
├── .opencode/                 # OpenCode extensions
│   ├── commands/              # Slash commands (e.g. /plan, /review-week)
│   ├── skills/                # Loadable skills (e.g. calendar management)
│   └── agents/                # Additional agent definitions
├── kanban/
│   └── board.md               # Task board with Backlog / In Progress / Done
└── state/                     # Durable context that persists across sessions
    └── (priorities, decisions, memory — format TBD)
```

## Config

The executive environment's config lives directly in `~/Executive/`:
- `opencode.json` — project-level config (overrides global)
- `AGENTS.md` — executive-specific system prompt and rules

Shared config (API keys, default model) comes from the global layer
at `~/.config/opencode/`. See `docs/02-profile-anatomy.md` for details.

## Design intent

- The executive assistant has access to `~/Executive/` (read/write)
- The executive assistant may have read access to `~/Vault/` (TBD)
- State files in `~/Executive/state/` carry across sessions
- The kanban board is the primary task-management surface
- Personal/sensitive data lives in `~/Vault/`, not `~/Executive/`

## Not yet decided

- Exact format and schema for state files
- Whether calendar is a file, an MCP integration, or both
- How memory/preferences persist (flat file vs structured)
- What MCP servers or tools the executive environment should use
- Whether the executive assistant gets read access to `~/Vault/`
