# System Overview

Ashley Cottrell's personal AI operating system.

## Purpose

A minimal, role-based system for managing life and building software,
operated through AI coding tools. The `orbit` CLI manages roles
(environments and workspaces), scaffolds directories, and launches
adapters (OpenCode, Cursor, etc.) in the correct directory.

## Orbit CLI

`orbit` is a Go CLI that manages this system. It knows two role types:

- **Environment** — a single directory that is itself the AI instance
  (e.g. ~/Executive/)
- **Workspace** — a parent directory whose children are each independent
  AI instances (e.g. ~/Dev/active/, where each project is its own env)

Adapters define what tool launches in each directory (OpenCode by default).
See `decisions/ADR-003-orbit-cli.md` for the full rationale.

## Domains

| Domain | Model | Home directory |
|--------|-------|---------------|
| Executive | Single long-lived OpenCode environment for life management | ~/Executive/ |
| Engineering | Launcher that opens OpenCode inside individual project directories | ~/Dev/active/<project>/ |

The executive assistant is one environment. You type `assistant` (alias
for `orbit open executive`), OpenCode opens in ~/Executive/, and you
manage your life.

The engineering assistant is many environments. You type `dev` (alias
for `orbit open engineering`), pick a project, and OpenCode opens scoped
entirely to that project's directory. Each project is self-contained.

## Top-level directory layout

| Path | Role |
|------|------|
| ~/AI/ | AI tooling and this architecture repo |
| ~/AI/architecture/ | Source of truth for system design + orbit CLI source (this repo) |
| ~/Executive/ | Executive assistant — single OpenCode environment |
| ~/Dev/ | Engineering — contains projects, not itself an OpenCode env |
| ~/Dev/active/ | Active engineering projects (each is its own OpenCode env) |
| ~/Vault/ | Sensitive and personal data (health, journal, financial) |
| ~/Archive/ | Timestamped archives of retired material |
| ~/.config/opencode/ | Global OpenCode config shared across all sessions |
| ~/.config/orbit/ | Orbit CLI config (roles, adapters, archive path) |

## Config layering

OpenCode merges config from multiple locations. We use two layers:

1. **Global** (`~/.config/opencode/`) — provider API keys, default model,
   shared agents, shared commands, shared skills, personal rules
2. **Project** (the directory OpenCode launches in) — project-specific
   config that overrides global for conflicts

The executive environment is one project: ~/Executive/.
Each engineering project is its own project: ~/Dev/active/<name>/.

## Principles

1. Roles are general. New domains require config changes, not code changes.
2. Domains are cleanly separated. Executive and engineering do not share state.
3. Each AI session is scoped to the directory it launched in.
4. Adapters are pluggable. OpenCode is the default, but any tool can be used.
5. File-based simplicity first. Local tools second. Heavy integrations only when justified.
6. The architecture repo is the single source of truth for system design.
7. Nothing is deleted. Retired material is archived with timestamps.
8. Sensitive data lives in ~/Vault/, never in the architecture repo.
