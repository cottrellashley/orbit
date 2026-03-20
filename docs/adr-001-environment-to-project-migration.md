# ADR-001: Additive Migration from Environment to Project

**Status:** Accepted  
**Date:** 2026-03-09  
**Authors:** Orbit team

---

## Context

Orbit's core registry entity is `domain.Environment`, representing a
registered filesystem path. The roadmap calls for richer project-level
metadata: repository topology detection (single-repo vs multi-repo),
auto-detected integration tags (git, python, uv, node, opencode, etc.),
and contained-repo metadata.

The term "project" better describes what users register. The rename also
aligns with the planned Orbit config workspace (`~/.config/orbit/`) and
GitHub adapter work.

## Decision

We adopt an **additive migration** strategy rather than a big-bang rename.

### Why additive

- **Zero breakage.** Existing CLI commands, HTTP API endpoints, adapter
  storage (jsonstore `environments.json`), and the web UI continue to
  work unchanged throughout the migration.
- **Incremental delivery.** Each agent/PR can ship independently: domain
  types first, then adapters, then driver UI, then cleanup.
- **Rollback safety.** If a later stage surfaces problems, earlier stages
  remain stable and deployed.

### Stages

| Stage | Scope | Breaking? | PR |
|-------|-------|-----------|----|
| **1 — Domain + Ports** | Add `domain.Project`, `ProjectTopology`, `IntegrationTag`, `RepoInfo`. Add `port.ProjectRepository`, `port.ConfigWorkspace`. Add bridge helpers (`ProjectFromEnvironment`, `EnvironmentFromProject`, `ProjectRepositoryFromEnvRepo`). Add `app.ProjectService` stub. | No | This PR |
| **2 — Native adapter** | Implement `ProjectRepository` in jsonstore (new file `projects.json` or unified storage). Implement `ConfigWorkspace` adapter for `~/.config/orbit/`. | No | — |
| **3 — App migration** | Migrate `SessionService`, `OpenService`, `DoctorService` to accept `ProjectRepository` (via bridge initially, native later). Add integration-detection logic. | No | — |
| **4 — Driver migration** | Add `/api/projects` endpoints. Update web UI. Add `orbit project` CLI commands alongside `orbit env`. | No | — |
| **5 — Deprecation** | Mark `Environment*` types, ports, and endpoints as deprecated. Log warnings on use. | No (soft) | — |
| **6 — Cleanup** | Remove `Environment`, `EnvironmentRepository`, `EnvironmentService`, old endpoints, old JSON file. | **Yes** | — |

### Compatibility guarantees

- Stages 1–4 are fully non-breaking. Existing behavior is untouched.
- Stage 5 adds deprecation warnings but does not remove functionality.
- Stage 6 is the only breaking change and will be gated behind a major
  version bump (or a clear migration window with tooling to convert
  `environments.json` → `projects.json`).

### Bridge mechanism

`port.ProjectRepositoryFromEnvRepo()` wraps any `EnvironmentRepository`
as a `ProjectRepository`, converting on every call via
`domain.ProjectFromEnvironment` / `domain.EnvironmentFromProject`. This
means:

- New app code can depend on `ProjectRepository` immediately.
- The composition root wires `ProjectRepositoryFromEnvRepo(envRepo)` —
  no new adapter needed yet.
- Topology and integration data are **not preserved** through the bridge
  (the environment schema has no fields for them). Once the native
  adapter (stage 2) lands, topology/integration data persists properly.

### End-state

After stage 6:

- `domain.Project` is the sole registry entity.
- `port.ProjectRepository` is the sole persistence port.
- `app.ProjectService` replaces `app.EnvironmentService`.
- The bridge helpers and `Environment` type are removed.
- Storage is `projects.json` (or equivalent).

## Consequences

- Temporary code duplication between Environment and Project paths.
- Bridge adapter adds a thin conversion layer with negligible perf cost.
- All new feature work (GitHub adapter, agent management, session
  redesign) should target `Project` / `ProjectRepository` from the start.
