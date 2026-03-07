# Executive System

## Purpose

The executive assistant manages Ashley's life: planning, task tracking,
calendar, priorities, decisions, and personal knowledge.

## Working directory

~/Executive/

## Current structure

- `kanban/board.md` — task board with Backlog / In Progress / Done lanes
- `state/` — durable context that persists across sessions (priorities, decisions, memory)

## OpenCode profile

~/AI/opencode/profiles/executive/

- `opencode.jsonc` — provider, model, agent, and MCP server configuration
- `AGENTS.md` — agent role description and behavioral instructions

## Design intent

- The executive profile should have access to ~/Executive/ and ~/Vault/ (read)
- State files in ~/Executive/state/ carry across sessions
- The kanban board is the primary task-management surface
- Personal/sensitive data lives in ~/Vault/, not ~/Executive/

## Not yet decided

- Exact format and schema for state files
- Whether calendar is a file, an MCP integration, or both
- How memory/preferences persist (flat file vs structured)
- What MCP servers or tools the executive profile should use
