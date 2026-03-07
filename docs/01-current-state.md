# Current State

Last updated: 2026-03-07

## What exists

### ~/AI/architecture/
- This repository. Bootstrapped 2026-03-07.
- Contains design docs, ADRs, and state tracking.
- One commit (bootstrap) + uncommitted updates reflecting the
  "profiles are directories" redesign.

### ~/Executive/
- `kanban/board.md` — empty kanban board (Backlog / In Progress / Done)
- `state/README.md` — placeholder describing purpose
- No OpenCode config yet (no `opencode.json`, `AGENTS.md`, or `.opencode/`)

### ~/Dev/
- `active/` — empty directory for active engineering projects
- No projects exist yet

### ~/Vault/
- Empty. Reserved for sensitive/personal data.

### ~/Archive/
- `pre-opencode-reset-2026-03-07-0039/` — full archive of previous setup:
  - `claude/` — previous Claude Code workspace (AshleyDB, CLAUDE.md, programs, repos)
  - `dev/` — previous dev directory
  - `repos/` — previous repos directory
  - `dotfiles/` — .opencode, .codex, .claude, .claude.json
  - `dotconfig/` — .config/opencode, .config/codex
- `obsolete-profiles-2026-03-07-1517/` — archived `~/AI/opencode/profiles/`
  (superseded by ADR-002: profiles are directories)

### ~/.config/opencode/
- Does NOT exist yet (was archived during reset)
- Needs to be created with global config (API keys, default model, shared config)

## What is configured

- OpenCode is installed
- GitHub CLI (gh) is installed and authenticated as cottrellashley
- No OpenCode config files exist anywhere (global or project-level)
- No MCP servers, plugins, or tools are currently active
- No git repos are initialized in the working directories (only in architecture)

## What is not yet decided

See state/open-questions.md
