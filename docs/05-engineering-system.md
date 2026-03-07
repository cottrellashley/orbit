# Engineering System

## Purpose

The engineering assistant helps Ashley write, review, debug, and manage
software projects.

## Working directory

~/Dev/

## Current structure

- `active/` — active engineering projects and repositories

## OpenCode profile

~/AI/opencode/profiles/engineer/

- `opencode.jsonc` — provider, model, agent, and MCP server configuration
- `AGENTS.md` — agent role description and behavioral instructions

## Design intent

- The engineering profile should have access to ~/Dev/ and relevant tool configs
- Each project lives as its own directory (typically a git repo) under ~/Dev/active/
- The engineer does not access ~/Executive/ or ~/Vault/

## Not yet decided

- Whether a project-scaffolding convention is needed
- What MCP servers or tools the engineering profile should use
- Whether build/CI tooling config lives here or in ~/AI/opencode/
