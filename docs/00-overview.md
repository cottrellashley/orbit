# System Overview

Ashley Cottrell's personal AI operating system.

## Purpose

A minimal, two-domain system for managing life and building software,
operated through OpenCode as the primary AI interface.

## Domains

| Domain | Model | Home directory |
|--------|-------|---------------|
| Executive | Single long-lived OpenCode environment for life management | ~/Executive/ |
| Engineering | Launcher that opens OpenCode inside individual project directories | ~/Dev/active/<project>/ |

The executive assistant is one environment. You type `assistant`, OpenCode
opens in ~/Executive/, and you manage your life.

The engineering assistant is many environments. You type `dev`, enter a
developer shell, pick a project, and OpenCode opens scoped entirely to
that project's directory. Each project is self-contained. OpenCode never
reaches outside the project it opened in.

## Top-level directory layout

| Path | Role |
|------|------|
| ~/AI/ | AI tooling and this architecture repo |
| ~/AI/architecture/ | Source of truth for system design (this repo) |
| ~/Executive/ | Executive assistant — single OpenCode environment |
| ~/Dev/ | Engineering — contains projects, not itself an OpenCode env |
| ~/Dev/active/ | Active engineering projects (each is its own OpenCode env) |
| ~/Vault/ | Sensitive and personal data (health, journal, financial) |
| ~/Archive/ | Timestamped archives of retired material |
| ~/.config/opencode/ | Global OpenCode config shared across all sessions |

## Config layering

OpenCode merges config from multiple locations. We use two layers:

1. **Global** (`~/.config/opencode/`) — provider API keys, default model,
   shared agents, shared commands, shared skills, personal rules
2. **Project** (the directory OpenCode launches in) — project-specific
   config that overrides global for conflicts

The executive environment is one project: ~/Executive/.
Each engineering project is its own project: ~/Dev/active/<name>/.

## Principles

1. Two domains, cleanly separated. Executive and engineering do not share state.
2. Each OpenCode session is scoped to the directory it launched in. It does not
   reach outside that directory.
3. File-based simplicity first. Local tools second. Heavy integrations only when justified.
4. The architecture repo is the single source of truth for system design.
5. Nothing is deleted. Retired material is archived with timestamps.
6. Sensitive data lives in ~/Vault/, never in the architecture repo.
