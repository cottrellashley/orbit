# ADR-002: Profiles are directories

**Date:** 2026-03-07
**Status:** Accepted

## Context

The initial design (ADR-001 era) assumed profiles would live in a
central `~/AI/opencode/profiles/` tree, with each profile containing
an `opencode.jsonc` and `AGENTS.md` that would somehow be referenced
when launching OpenCode.

After researching OpenCode's actual configuration system, we discovered
that OpenCode's config is **directory-based**. When `opencode` runs in
a directory, it loads `opencode.json`, `AGENTS.md`, and `.opencode/`
from that directory (traversing up to the git root). Config merges:
global (`~/.config/opencode/`) + project (the launch directory), with
project overriding global.

This means:
- A "profile" is simply a directory with its own OpenCode configuration.
- There is no separate profile registry or selection mechanism.
- The `~/AI/opencode/profiles/` tree served no purpose.

## Decision

1. A profile is a directory. The executive profile is `~/Executive/`.
   Each engineering profile is `~/Dev/active/<project>/`.
2. OpenCode config files (`opencode.json`, `AGENTS.md`, `.opencode/`)
   live directly in the project directory, not in a central profiles tree.
3. Shared config goes in the global layer (`~/.config/opencode/`).
4. The obsolete `~/AI/opencode/profiles/` directory is archived.

## Consequences

- The executive environment is fully self-contained in `~/Executive/`.
- Each engineering project is fully self-contained in its own directory.
- No launcher needs to copy or symlink config files.
- The `assistant` alias simply runs `opencode` in `~/Executive/`.
- The `dev` alias opens a shell in `~/Dev/` where the user picks a project,
  then runs `opencode` inside that project's directory.
- `~/AI/opencode/profiles/` has been archived to
  `~/Archive/obsolete-profiles-2026-03-07-1517/`.
