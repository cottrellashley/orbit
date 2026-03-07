# Project Model

This document describes how engineering projects are structured and
managed under the `dev` domain.

## Core concept

`~/Dev/` is NOT an OpenCode environment. It is a container for projects.
Each project under `~/Dev/active/<name>/` is its own fully self-contained
OpenCode environment. OpenCode never reaches outside the project directory.

## Project lifecycle

### Create
A new project is created as a directory under `~/Dev/active/`:
```
~/Dev/active/<name>/
├── opencode.json          # project-specific OpenCode config (optional)
├── AGENTS.md              # project-specific agent rules (optional)
├── .opencode/             # project-specific commands/skills/agents (optional)
│   ├── commands/
│   ├── skills/
│   └── agents/
└── <project files>        # source code, configs, etc.
```

Most projects will be git repos. The OpenCode config files can be
committed to the repo or kept local (via `.gitignore`).

### Open
To work on an existing project:
1. Type `dev` to enter the developer shell
2. Select a project from `~/Dev/active/`
3. OpenCode launches scoped to that project's directory

### Archive
When a project is no longer active:
```
mv ~/Dev/active/<name> ~/Archive/<name>-<timestamp>
```

### Delete
Projects are never deleted. They are archived.

## Isolation

Each project is fully isolated:
- OpenCode loads config only from the project directory + global config
- No project can access another project's files
- No project can access `~/Executive/` or `~/Vault/`
- The global config layer provides shared defaults, but projects override

## Project scaffolding

Whether we use a scaffolding template for new projects is an open
question. Options:
1. No template — each project starts from scratch or from `git clone`
2. A minimal template with a default `opencode.json` and `AGENTS.md`
3. A `dev new <name>` command that scaffolds a project from a template

This is recorded in `state/open-questions.md`.
