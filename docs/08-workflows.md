# Workflows

## Starting an executive session

```
$ assistant
```

1. The `assistant` alias runs `opencode` in `~/Executive/`
2. OpenCode loads `~/Executive/opencode.json` + `AGENTS.md` + `.opencode/`
3. Global config from `~/.config/opencode/` provides API keys and defaults
4. The agent reads `~/Executive/state/` for persistent context
5. Work proceeds against `~/Executive/` (kanban, planning, etc.)

## Starting an engineering session

```
$ dev
```

1. The `dev` alias opens a developer shell in `~/Dev/`
2. The shell presents active projects from `~/Dev/active/`
3. User selects a project (or creates a new one)
4. `opencode` launches inside `~/Dev/active/<project>/`
5. OpenCode loads the project's config + global config
6. Work proceeds scoped to that project

## Architecture planning

1. Open terminal
2. Launch OpenCode in `~/AI/architecture/` (or use any session)
3. Read `docs/00-overview.md`, `docs/01-current-state.md`, `docs/09-editing-protocol.md`
4. Follow the editing protocol: propose changes, get approval, then edit
5. Commit changes to the architecture repo

## Not yet decided

- Exact implementation of the `assistant` alias (simple alias vs script)
- Exact implementation of the `dev` alias (shell script with project picker)
- Whether tmux is used for session management
- Session handoff or continuity mechanisms
- How the kanban board is updated (manual vs agent-driven)
