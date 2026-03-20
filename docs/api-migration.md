# API Migration: Environments to Projects

**Status:** Stage 5 (UI) complete (ADR-001)
**Date:** 2026-03-09

## Overview

Orbit is migrating from "Environment" to "Project" as its core registry
entity. This document describes the HTTP API and CLI changes introduced
in Stage 4 and the recommended client migration path.

## New Project Endpoints

These are the preferred endpoints for all new integrations.

| Method   | Path                    | Description                | Status Code |
|----------|-------------------------|----------------------------|-------------|
| `GET`    | `/api/projects`         | List all projects          | 200         |
| `GET`    | `/api/projects/{name}`  | Get a single project       | 200 / 404   |
| `POST`   | `/api/projects`         | Register a new project     | 201 / 400   |
| `DELETE` | `/api/projects/{name}`  | Remove a project           | 200 / 404   |

### Project response shape

```json
{
  "name": "my-project",
  "path": "/home/user/my-project",
  "description": "A description",
  "profile_name": "",
  "topology": "single-repo",
  "integrations": ["git", "python"],
  "repos": [
    {
      "path": "/home/user/my-project",
      "remote_url": "git@github.com:user/repo.git",
      "current_branch": "main"
    }
  ],
  "created_at": "2026-03-09T00:00:00Z",
  "updated_at": "2026-03-09T00:00:00Z"
}
```

Project responses include `topology`, `integrations`, and `repos` fields
that are not available on the legacy Environment responses.

### POST request body

```json
{
  "name": "my-project",
  "path": "/home/user/my-project",
  "description": "optional"
}
```

Both `name` and `path` are required. `description` is optional.

## Legacy Environment Endpoints (Still Supported)

These endpoints continue to work exactly as before. No behavioral changes.

| Method   | Path                          | Description              | Status Code |
|----------|-------------------------------|--------------------------|-------------|
| `GET`    | `/api/environments`           | List all environments    | 200         |
| `POST`   | `/api/environments`           | Register an environment  | 201 / 400   |
| `DELETE` | `/api/environments/{name}`    | Remove an environment    | 200 / 404   |

Environment and Project registries are **independent stores** during the
migration period. Creating a project does not create an environment and
vice versa. After Stage 6 (cleanup), only the project endpoints will
remain.

## CLI Commands

### New: `orbit project`

```
orbit project list              # List registered projects
orbit project add <name> <path> # Register a project
orbit project show <name>       # Show project details
orbit project remove <name>     # Remove a project (alias: rm)
```

The `orbit project` command group (alias: `orbit proj`) is the preferred
way to manage registrations going forward.

### Existing: `orbit serve`, `orbit attach`

These commands are unchanged. `orbit serve` now serves both `/api/projects`
and `/api/environments` routes.

## Shared Endpoints (Unchanged)

These endpoints are unaffected by the migration:

| Method   | Path                              | Description                |
|----------|-----------------------------------|----------------------------|
| `GET`    | `/api/servers`                    | List discovered servers    |
| `GET`    | `/api/sessions`                   | List all sessions          |
| `GET`    | `/api/sessions/{id}`              | Get a session              |
| `POST`   | `/api/sessions/{id}/abort`        | Abort a session            |
| `DELETE` | `/api/sessions/{id}`              | Delete a session           |
| `GET`    | `/api/doctor`                     | Run doctor checks          |
| `*`      | `/api/proxy/{port}/{path...}`     | Proxy to OpenCode servers  |

## Error Response Shape

All endpoints use a consistent error format:

```json
{
  "error": "human-readable error message"
}
```

HTTP status codes follow standard conventions: 400 for validation errors,
404 for not-found, 500 for internal errors.

## Recommended Client Migration Path

1. **Now (Stage 4):** Start using `/api/projects` for new integrations.
   Continue using `/api/environments` for existing functionality.
2. **Stage 5 (future):** Environment endpoints will emit deprecation
   warnings in response headers. Begin migrating remaining callers.
3. **Stage 6 (future, breaking):** Environment endpoints will be removed.
   All callers must use `/api/projects` by this point.

## Planned Future Deprecation

The following will be deprecated in Stage 5 and removed in Stage 6:

- `GET /api/environments`
- `POST /api/environments`
- `DELETE /api/environments/{name}`
- `domain.Environment` type
- `port.EnvironmentRepository` interface
- `app.EnvironmentService`

No breaking changes will be made without a major version bump.

## Stage 5: Web UI Migration (Complete)

The embedded web UI (`internal/driver/server/static/index.html`) has been
updated to reflect the Environment-to-Project migration.

### What was migrated

| Area | Before | After |
|------|--------|-------|
| Sidebar tagline | "AI Environment Manager" | "AI Project Manager" |
| Primary nav item | "Environments" | "Projects" |
| Dashboard summary card | "Environments registered" | "Projects registered" |
| Dashboard section | shows Environments table | shows Project cards |
| Add modal (primary) | "Add Environment" via `/api/environments` | "Add Project" via `/api/projects` |
| Detail page | none | Project detail with topology, integrations, repos |
| API endpoint used | `GET /api/environments` | `GET /api/projects`, `GET /api/projects/{name}` |
| Sessions table column | "Environment" | "Project" |

### What was added (new features)

- **Projects page**: Card-based grid showing name, description, path,
  topology badge, and integration tags. Click to open project detail.
- **Project detail page**: Full detail view with metadata grid, integration
  tags (color-coded per tech), and repository table (path, remote, branch).
- **Jump/Launch UX**: Every session row has terminal (TUI) and globe (Web)
  action buttons. TUI copies `opencode attach` command to clipboard. Web
  opens the OpenCode server in a new browser tab.
- **Assistant panel (stub)**: Dedicated "Assistant" nav item under System.
  Shows a placeholder with "Coming soon" status. Input area is disabled.
  Ready for backend integration when the Orbit assistant server is built.
- **Markdown renderer**: `renderMarkdown()` function converts CommonMark
  subset (headings, bold, italic, code blocks, inline code, links, lists,
  blockquotes, horizontal rules) to styled HTML. Available for any panel
  that needs to render markdown content.
- **Toast notifications**: `showToast()` utility for non-blocking feedback
  (used by TUI clipboard copy action).
- **Integration tag styles**: Color-coded CSS classes for git, python, uv,
  node, opencode, go, rust, docker.

### What is preserved (backward compatibility)

- **Environments (Legacy) page**: Still present in sidebar under "Manage",
  labeled "Environments". Shows a deprecation banner directing users to
  Projects. Full CRUD still works via `/api/environments`.
- **Legacy Add Environment modal**: Still accessible from the Environments
  page, with deprecation notice.
- **All legacy API calls**: The UI continues to fetch from `/api/environments`
  for the legacy page. No environment endpoints were removed.

### What is stubbed (pending backend work)

- **Orbit Assistant**: UI shell is complete (nav item, chat container,
  disabled input). Requires a dedicated OpenCode server backend to become
  functional. Expected contract: the assistant will be an OpenCode session
  proxied through Orbit's API, similar to existing session message fetching.
- **Markdown panel**: Renderer is built but not yet wired to any data
  source. Will be used when agent plans, skills, and profile descriptions
  become browsable in the UI.

### Expected API contracts for future features

**Assistant (when backend is ready):**
- `POST /api/assistant/message` — send a message to the Orbit assistant
- `GET /api/assistant/messages` — fetch assistant conversation history
- Response shape: compatible with existing session message format

**Markdown content browsing:**
- `GET /api/projects/{name}/files/{path...}` — read files from project dir
- Response: `{ "content": "...", "path": "...", "type": "markdown" }`
