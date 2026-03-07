# System Overview

Ashley Cottrell's personal AI operating system.

## Purpose

A minimal, two-domain system for managing life and building software,
operated through OpenCode as the primary AI interface.

## Domains

| Domain | Purpose | Home directory |
|--------|---------|---------------|
| Executive | Life management, planning, task coordination, decision support | ~/Executive/ |
| Engineering | Software development, code authoring, architecture, project management | ~/Dev/ |

## Top-level directory layout

| Path | Role |
|------|------|
| ~/AI/ | AI tooling, profiles, and this architecture repo |
| ~/AI/architecture/ | Source of truth for system design (this repo) |
| ~/AI/opencode/ | OpenCode installation, profiles, and scripts |
| ~/Executive/ | Executive assistant working state (kanban, durable state) |
| ~/Dev/ | Engineering projects and active work |
| ~/Vault/ | Sensitive and personal data (health, journal, financial) |
| ~/Archive/ | Timestamped archives of previous setups and retired material |

## Principles

1. Two domains, cleanly separated. Executive and engineering do not share state.
2. File-based simplicity first. Local tools second. Heavy integrations only when justified.
3. The architecture repo is the single source of truth for system design.
4. Nothing is deleted. Retired material is archived with timestamps.
5. Sensitive data lives in ~/Vault/, never in the architecture repo.
