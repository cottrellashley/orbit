# Profile Anatomy

A "profile" is just a directory that OpenCode launches in. There is no
separate profile registry. The directory's config files define the
behavior of the OpenCode session.

## Config files in a project directory

| File / directory | Purpose |
|---|---|
| `opencode.json` | Provider, model, agent definitions, MCP servers |
| `AGENTS.md` | System prompt / rules for the default agent |
| `.opencode/commands/*.md` | Slash commands (markdown with frontmatter) |
| `.opencode/skills/<name>/SKILL.md` | On-demand loadable skills |
| `.opencode/agents/*.md` | Additional agent definitions |

All of these are optional. If absent, OpenCode falls back to global
config or built-in defaults.

## Config layering

OpenCode merges config from two locations:

1. **Global** — `~/.config/opencode/`
   - Provider API keys
   - Default model
   - Shared agents, commands, skills
   - Personal rules (`instructions` array)
2. **Project** — the directory OpenCode launches in
   - Project-specific overrides
   - Project-specific agents, commands, skills
   - Project-specific `AGENTS.md` system prompt

Project config overrides global config where they conflict.

## What goes where

| Content | Location |
|---|---|
| API keys and provider config | Global (`~/.config/opencode/config.json`) |
| Default model preference | Global |
| Commands/skills useful everywhere | Global (`.config/opencode/.opencode/`) |
| Executive-specific behavior | `~/Executive/` (opencode.json, AGENTS.md, .opencode/) |
| Project-specific behavior | `~/Dev/active/<name>/` (opencode.json, AGENTS.md, .opencode/) |

## Environment variable

`OPENCODE_CONFIG_DIR` can override the global config location. We use
the default (`~/.config/opencode/`) unless there is a reason to change it.

## CLI flags

OpenCode supports launch-time overrides:
- `opencode --model <model>` — override the model for this session
- `opencode --agent <agent>` — start with a specific agent

These are useful for one-off overrides but should not replace config files
for persistent preferences.
