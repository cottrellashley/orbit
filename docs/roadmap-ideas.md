# Orbit Roadmap & Ideas

Captured 2026-03-09. Raw brain dump — not prioritized, not designed yet.

---

## Core Philosophy Reminder

Orbit is a **manager**, NOT an AI agentic tool. It manages OpenCode instances,
projects, profiles, and integrations. Users should be able to **jump into**
sessions, agents, repos, etc. — Orbit is the control plane, not the execution
plane.

---

## 0. Orbit Assistant — The Meta AI (HIGHEST PRIORITY)

Orbit runs its OWN dedicated OpenCode server instance — a built-in chatbot
that is the user's guide to their entire setup. This is NOT a general-purpose
coding agent. It is a **small wrapper chatbot around an OpenCode server**
purpose-built for Orbit management.

### What it is

A fully contextualised assistant that:
- Has **full knowledge of the user's Orbit state** — projects, profiles,
  servers, sessions, agents, installed tools, doctor results, everything
- Has **memory** — remembers past interactions, user preferences, setup
  history across sessions
- Can **read and act on Orbit state** — call Orbit's own API/domain to
  inspect config, check what's installed, list projects, etc.
- Knows **best practices** for project setup, profile configuration, agent
  configuration, MCP server setup, GitHub integration, etc.

### What it does

- **Onboarding**: Walks new users through first-time setup end to end
- **Installation help**: Detects missing tools (uv, gh, opencode, git) and
  guides the user through installing them step by step
- **Setup assistance**: Helps create projects, configure profiles, set up
  MCP servers, write skills and slash commands
- **Troubleshooting**: When doctor checks fail, the assistant explains why
  and walks the user through the fix
- **Best practices**: Advises on project structure, profile design, agent
  configuration — opinionated guidance from a knowledgeable helper
- **State queries**: "What projects do I have?", "Is uv installed?",
  "Which agents are configured for project X?" — the assistant can answer
  by querying Orbit's domain directly

### Architecture

- Orbit launches a **dedicated OpenCode server** just for itself (separate
  from user project servers) on `orbit serve` startup
- The assistant is a **sub-agent** with a system prompt containing Orbit
  domain knowledge, best practices, and tool access to Orbit's own APIs
- Exposed in the Web UI as a chat panel (always accessible, like a help
  sidebar or dedicated page)
- Also accessible via TUI and CLI drivers
- The OpenCode server gives it full LLM capabilities — the Orbit-specific
  context and tools make it a specialised management assistant
- State/memory persists across sessions via the OpenCode server's own
  session storage

### Key constraint

This is a SMALL wrapper. Orbit doesn't reimplement an LLM runtime — it
spins up an OpenCode server and gives it the right context, tools, and
system prompt. The intelligence comes from OpenCode; the specialisation
comes from Orbit's domain knowledge piped in.

---

## 1. Rename "Environments" to "Projects"

The term "environment" is wrong. These are **projects**. A project is a
workspace that may:

- BE a git repo (single-repo project)
- CONTAIN multiple git repos (monorepo or multi-repo project)

When a user clicks on a project, we detect which case it is:
- If it IS a repo: show repo info directly
- If it CONTAINS repos: show a dropdown/expandable list of contained repos

### Project Integration Tags

Each project should show small visual tags/icons based on detected tooling:
- OpenCode (if configured)
- Python logo (if python project)
- uv logo (if uv detected)
- Git logo (if git repo)
- Node/npm/etc. as applicable

These tags are auto-detected from the project directory contents.

---

## 2. Agent Management Page

A dedicated page for managing **OpenCode agents**:
- List all configured agents
- Create new agents
- Configure existing agents (model, system prompt, tools, etc.)
- Delete agents
- View agent status/health

This is about setting up and configuring the agents that OpenCode will use,
not about running them directly.

---

## 3. Session View Redesign — Hierarchical Sessions

The current sessions view is flat and too busy. Fix:

- Top level shows only **individual sessions** (parent sessions)
- Each session row is expandable (dropdown/accordion)
- Expanding reveals **sub-sessions** (sub-agents spawned within that session)
- Collapsed by default — clean and scannable
- Expand to see the full tree when needed

---

## 4. Session Launch — Jump Into Sessions

Users should be able to **open/launch** a session and choose:
- **OpenCode TUI**: opens terminal with `opencode attach` to that session
- **OpenCode Web**: redirects to the OpenCode web UI for that session

Jump straight INTO the session from the Orbit UI. Orbit is the launchpad.

---

## 5. Profile Builder

A full profile creation/editing environment where users can:
- Build new **Skills** (custom skill definitions)
- Configure new **MCP servers** (tool servers)
- Create new **Slash commands**
- Set model preferences, system prompts, tool access

Then users can **create projects using those profiles** — a profile is a
reusable configuration template that gets applied when setting up a project.

---

## 6. GitHub Adapter / Port

A new adapter implementing a GitHub port interface. Must go through
`domain/ -> port/ -> adapter/` (hexagonal, no shortcuts). Capabilities:

- **Auth**: GitHub authentication (token-based, OAuth device flow?)
- **Auth caching**: Persist and refresh tokens
- **Repo listing**: List user's repos
- **Repo issue lists**: Fetch issues for a repo
- **GitHub Agent management**: Configure/view GitHub Copilot agents or
  GitHub-integrated AI agents
- **Redirect/jump**: From Orbit UI, jump directly into a GitHub repo,
  issue, PR, agent, etc.

All via core domain/app so the same functionality is available through:
- TUI driver
- Web UI driver
- CLI driver

---

## 7. Enhanced Doctor Checks

Add these checks to the doctor system:

- **uv installation check**: Is `uv` installed and on PATH?
- **gh installation check**: Is the GitHub CLI (`gh`) installed?
- **opencode installation check**: Is `opencode` installed and which version?
- **git repo detection**: Per-project — is it a repo? Does it contain repos?

These feed into project setup detection and the integration tags.

---

## 8. Navigation & Linking

Since Orbit is a manager/control plane, heavy emphasis on **jumping out**:
- Jump into GitHub agents (redirect to GitHub)
- Jump into repos (open in browser or terminal)
- Jump into sessions (TUI or web)
- Jump into issues, PRs, etc.

Orbit doesn't replicate these UIs — it links to them and launches them.

---

## 9. Markdown Rendering Engine (from OpenDoc)

Take the markdown rendering engine from **OpenDoc** (`cottrellashley/opendoc`)
— our own static site generator with integrated workbench and terminal — and
use it as Orbit's built-in markdown renderer. Many things in Orbit are really
just markdown files on disk — they should render beautifully in the UI when
a user clicks on them.

### What gets rendered

- **Agent Plans**: The set of markdown files an agent uses to plan, track
  findings, and maintain state. When a user clicks on an agent's plan, they
  see the rendered markdown — not raw text.
- **Skills**: When browsing/editing profiles, the skills (which are markdown
  definitions on disk) render nicely in-place.
- **Any markdown-based config**: READMEs, project docs, profile descriptions,
  slash command help text — anything stored as `.md` gets first-class
  rendering.

### How it works

- Reuse OpenDoc's rendering engine (don't reinvent)
- Render inline in the Orbit UI — clicking a component that is backed by a
  markdown file opens a rendered view, not a raw file dump
- Support code blocks with syntax highlighting, tables, headings, lists,
  links — the full CommonMark spec plus whatever OpenDoc already supports
- Read-only rendering in the browser; editing happens via the Orbit
  Assistant or external editor

---

## 10. Orbit Config Directory — Managed Workspace

A dedicated, closed directory specifically for Orbit's own configuration
and state. This is the user's space for managing their entire setup,
and the Orbit Assistant operates within it.

### Structure (something like `~/.config/orbit/` or `~/.orbit/`)

- `profiles/` — profile definitions (markdown + config)
- `skills/` — skill definitions (markdown)
- `agents/` — agent configurations
- `plans/` — agent plans, findings, tracking state (markdown)
- `mcp/` — MCP server configurations
- `commands/` — custom slash command definitions
- `state/` — runtime state (servers.json, session cache, etc.)

### Key properties

- **Browsable in the UI**: The Orbit web UI and TUI can navigate this
  directory structure, rendering markdown files inline (see section 9)
- **Orbit Assistant has full access**: The personalised Orbit agent can
  read, write, and organise files in this directory — helping users
  manage their setup through conversation
- **Self-contained**: Everything Orbit needs to know about the user's
  configuration lives here. Portable, backupable, version-controllable.
- **Not project data**: This is Orbit's own config space, separate from
  the user's project directories. Projects reference profiles from here,
  but the project code lives elsewhere.

---

## 11. UI Inspiration — Study OpenClaw

Take inspiration from **OpenClaw** (openclaw/openclaw) for UI/UX decisions —
specifically tab groupings, naming conventions, navigation structure, and
information hierarchy. NOT the implementation.

### Why OpenClaw

- Large established userbase — the UX has been battle-tested and iterated
- Reasonable to assume their tab groupings, naming, and navigation patterns
  reflect good user experience intuition loops
- Users coming from OpenClaw will find familiar patterns (Jakob's Law)

### What to study

- How they group features into tabs/sections
- What they name things (labels, menu items, section headers)
- Navigation flow — how users move between views
- Information density — what they show at a glance vs behind a click
- Empty states, onboarding, first-run experience

### What NOT to take

- Implementation details — we have our own architecture
- Visual design — we have our own design system (see UI redesign)
- Technology choices — Orbit is Go + single HTML, not whatever OpenClaw uses

Study the UX patterns, apply them where they make sense for Orbit's
management-focused use case.
