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
