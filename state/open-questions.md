# Open Questions

Items that need decisions before the architecture can advance.

## Executive system
- What is the schema/format for durable state files in `~/Executive/state/`?
- How should calendar be handled? (file, MCP integration, or both)
- How do memory and preferences persist across sessions?
- What MCP servers or tools should the executive environment use?
- Should the executive assistant have read access to `~/Vault/`?

## Engineering system
- Should there be a project-scaffolding template for new projects?
  (no template / minimal template / `dev new` command)
- What MCP servers or tools should engineering projects use?
- Should common engineering commands/skills go in global config or
  be copied per-project?

## Cross-cutting
- How are domain access boundaries enforced? (convention only, or tooling)
- Exact implementation of `assistant` alias (simple alias vs script)
- Exact implementation of `dev` alias (shell script with project picker)
- Whether tmux is used for session management
- What internal structure should `~/Vault/` have?
- What should be migrated (if anything) from the archived AshleyDB?

## Resolved
- ~~What is the exact invocation method for each profile?~~ —
  Profiles are directories. `assistant` = opencode in ~/Executive/.
  `dev` = pick a project, opencode in ~/Dev/active/<name>/.
  (ADR-002)
- ~~Should ~/AI/opencode/bin/ contain launcher scripts?~~ —
  No. Profiles are directories, not scripts. Simple shell aliases
  or small scripts in ~/.config/ or similar. (ADR-002)
