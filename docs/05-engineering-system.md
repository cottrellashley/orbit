# Engineering System

## Purpose

The engineering assistant helps Ashley write, review, debug, and manage
software projects.

## Model

A launcher that opens OpenCode inside individual project directories.
Type `dev`, enter a developer shell, pick or create a project, and
OpenCode opens scoped entirely to that project's directory.

`~/Dev/` itself is NOT an OpenCode environment. Each project under
`~/Dev/active/<name>/` is its own fully self-contained OpenCode
environment.

## Working directory

`~/Dev/active/<name>/` (one per project)

## Directory structure

```
~/Dev/
└── active/
    └── <project-name>/
        ├── opencode.json          # project-specific OpenCode config (optional)
        ├── AGENTS.md              # project-specific agent rules (optional)
        ├── .opencode/             # project-specific commands/skills/agents
        │   ├── commands/
        │   ├── skills/
        │   └── agents/
        └── <project files>        # source code, configs, etc.
```

## Config

Each project's config lives directly in its own directory:
- `opencode.json` — project-level config (overrides global)
- `AGENTS.md` — project-specific system prompt and rules

Shared config (API keys, default model) comes from the global layer
at `~/.config/opencode/`. See `docs/02-profile-anatomy.md` for details.

## Project lifecycle

See `docs/03-project-model.md` for the full lifecycle (create, open,
archive). Projects are never deleted, only archived.

## Isolation

- Each project is fully isolated from every other project
- No project can access `~/Executive/` or `~/Vault/`
- OpenCode loads config only from the project directory + global config

## Not yet decided

- Whether a project-scaffolding convention or template is needed
- What MCP servers or tools should be available to engineering projects
- Whether common engineering commands/skills go in global config
  or are copied per-project
