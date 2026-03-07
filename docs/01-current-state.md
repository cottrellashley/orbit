# Current State

Last updated: 2026-03-07

## What exists

### ~/AI/architecture/
- This repository. Bootstrapped 2026-03-07.
- Contains design docs, ADRs, state tracking, AND the orbit CLI source code.
- Go module: `github.com/cottrellashley/orbit`
- `orbit` binary builds and runs (6 commands: init, open, new, list, archive, status)
- No orbit config file exists yet (`~/.config/orbit/config.yaml`)

### ~/Executive/
- `kanban/board.md` — empty kanban board (Backlog / In Progress / Done)
- `state/README.md` — placeholder describing purpose
- No OpenCode config yet (no `opencode.json`, `AGENTS.md`, or `.opencode/`)
- Not yet registered as an orbit role

### ~/Dev/
- `active/` — empty directory for active engineering projects
- No projects exist yet
- Not yet registered as an orbit workspace

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

### ~/.config/orbit/
- Does NOT exist yet
- Needs to be created with initial config (adapters, roles)

## What is configured

- OpenCode is installed
- Go 1.25.7 is installed
- GitHub CLI (gh) is installed and authenticated as cottrellashley
- orbit CLI builds from source (not yet installed to PATH)
- No OpenCode config files exist anywhere (global or project-level)
- No orbit config file exists yet
- No MCP servers, plugins, or tools are currently active
- No git repos are initialized in the working directories (only in architecture)

## What is not yet decided

See state/open-questions.md
