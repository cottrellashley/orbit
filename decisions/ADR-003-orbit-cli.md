# ADR-003: Orbit CLI with adapter architecture

**Date:** 2026-03-07
**Status:** Accepted

## Context

The architecture design was complete in documentation but had no
executable tooling. Managing roles (environments and workspaces),
scaffolding directories, and launching AI coding tools was going to
require shell aliases and manual setup.

We considered three approaches:
1. Shell aliases and scripts — simple but fragile, no validation
2. A hardcoded CLI with executive/engineering baked in — too rigid
3. A general role-based CLI with pluggable adapters — flexible, minimal

Option 2 was rejected because it would require code changes to add
new domains (e.g. "research", "writing"). Option 1 was rejected
because it provides no lifecycle management (scaffolding, archiving).

## Decision

Build `orbit`, a Go CLI that:

1. **Manages roles** — named references to directories with a type
   (environment or workspace), optional tags, and an adapter.
2. **Scaffolds directories** — creates directory structure and config
   files when initializing roles or creating workspace projects.
3. **Launches via adapters** — an adapter is a named command (e.g.
   "opencode", "cursor") that orbit execs in the target directory.
   The CLI has no knowledge of what the adapter does.

### Core domain types

- **Role** — name, type (environment|workspace), path, adapter, tags
- **Adapter** — name, command, args, default flag
- **Environment** — single directory, single instance
- **Workspace** — parent directory, each child is an independent instance

### Commands

- `orbit init` — register a role, scaffold directory
- `orbit open` — launch adapter in a role's directory
- `orbit new` — create a project in a workspace
- `orbit list` — list roles with optional tag filter
- `orbit archive` — archive a role or workspace project
- `orbit status` — show config and role details

### Config

`~/.config/orbit/config.yaml` defines adapters and roles. The CLI
reads/writes this file. Roles reference directories; the CLI does not
own or manage the contents of those directories beyond initial scaffold.

## Consequences

- Adding a new domain is `orbit init <name> --type <type> --path <dir>` — zero code changes
- Adding a new adapter is a config entry — zero code changes
- Shell aliases become one-liners: `alias assistant="orbit open executive"`
- The architecture repo (`~/AI/architecture/`) is now also the Go source repo
- Architecture docs remain in `docs/`, `decisions/`, `state/` alongside the Go code
