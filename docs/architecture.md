# Orbit Architecture

Orbit uses a **hexagonal (ports & adapters) architecture**. This document
describes the layers, dependency rules, and conventions enforced throughout
the codebase.

## Layer Diagram

```
                         Drivers (primary adapters)
                    ┌──────────┬──────────┬──────────┐
                    │   CLI    │   TUI    │  HTTP    │
                    │  (cobra) │(bubbletea│  Server  │
                    │          │  — TBD)  │          │
                    └────┬─────┴────┬─────┴────┬─────┘
                         │          │          │
                         ▼          ▼          ▼
                    ┌──────────────────────────────────┐
                    │        Application Services       │
                    │  SessionService, EnvironmentService│
                    │  DoctorService, OpenService, ...  │
                    └──────────┬───────────────────────┘
                               │
                    ┌──────────▼───────────────────────┐
                    │     Ports (interfaces only)       │
                    │  SessionProvider, ServerLifecycle, │
                    │  EnvironmentRepository, ...       │
                    └──────────┬───────────────────────┘
                               │
                    ┌──────────▼───────────────────────┐
                    │     Domain (pure business logic)  │
                    │  Session, Server, Environment,    │
                    │  Profile, Report, OpenPlan, ...   │
                    └──────────────────────────────────┘

                    ┌──────────────────────────────────┐
                    │    Adapters (secondary adapters)   │
                    │  opencode, jsonstore, tmux        │
                    └──────────────────────────────────┘
```

## Package Layout

```
internal/
├── domain/          # Core business entities, zero external dependencies
├── port/            # Interfaces (driven ports) — depend only on domain
├── app/             # Application services — orchestrate domain + ports
├── adapter/         # Driven adapter implementations
│   ├── opencode/    #   SessionProvider + ServerLifecycle (OpenCode SDK)
│   ├── jsonstore/   #   EnvironmentRepository (JSON file persistence)
│   └── tmux/        #   WorkspaceManager (tmux subprocess)
└── driver/          # Driving adapters (primary adapters)
    ├── cli/         #   Cobra commands — includes composition root
    ├── server/      #   HTTP API + embedded web UI
    └── tui/         #   Bubbletea TUI (not yet implemented)
```

## Dependency Rules

These rules are absolute. Every code change must respect them.

### 1. Domain depends on NOTHING except the standard library

`internal/domain/` imports only `"errors"`, `"fmt"`, `"time"`, etc.
No third-party packages. No other internal packages.

Domain types must not carry serialization tags (JSON, YAML, etc.).
Serialization format is an adapter/driver concern, handled via DTOs.

### 2. Ports depend only on domain

`internal/port/` defines interfaces whose method signatures reference only
`domain` types and stdlib types (`context.Context`, `error`, etc.).

### 3. App services depend on ports and domain

`internal/app/` imports `internal/port` and `internal/domain`.
App services **never** import adapters or drivers.

When one app service needs another, it depends on a consumer-defined
interface (see `sessionQuerier` in `app/open.go`), not on the concrete
service type.

### 4. Adapters depend on ports, domain, and external libraries

`internal/adapter/*/` imports `internal/port`, `internal/domain`, and any
third-party SDK (e.g., `github.com/sst/opencode-sdk-go`). Each adapter
implements one or more port interfaces.

Adapters own their serialization DTOs. For example:
- `jsonstore` has `envDTO` with JSON tags, and converts to/from
  `domain.Environment`.
- `opencode/state.go` has `managedServerDTO` with JSON tags, and
  converts to/from `domain.ManagedServer`.

### 5. Drivers depend on app services (via interfaces) and domain

Drivers (`cli`, `server`, `tui`) call app service methods and reference
domain types for data.

The HTTP server defines its own consumer-side interfaces
(`server.SessionService`, `server.EnvironmentService`, etc.) so it
never imports `internal/app` directly.

### 6. The composition root is the ONLY place drivers touch adapters

`internal/driver/cli/root.go` is the composition root. It is the single
location where adapters are imported and wired into app services.

Adapter types must not leak beyond factory functions in the composition
root. Command logic (e.g., `newServeCmd`) references only port interfaces
and app services.

## Ports Reference

| Port | File | Implemented By |
|------|------|----------------|
| `EnvironmentRepository` | `port/repository.go` | `jsonstore.Store` |
| `ProfileRepository` | `port/repository.go` | *not yet implemented* |
| `SessionProvider` | `port/session_provider.go` | `opencode.Adapter` |
| `ServerLifecycle` | `port/server_lifecycle.go` | `opencode.ServerManager` |
| `WorkspaceManager` | `port/workspace.go` | `tmux.Manager` |

## App Services Reference

| Service | File | Dependencies |
|---------|------|-------------|
| `SessionService` | `app/session.go` | `EnvironmentRepository`, `SessionProvider`, `ServerLifecycle` (optional) |
| `EnvironmentService` | `app/environment.go` | `EnvironmentRepository`, `ProfileRepository` |
| `DoctorService` | `app/doctor.go` | `EnvironmentRepository`, `SessionProvider`, `ProfileRepository` |
| `OpenService` | `app/open.go` | `EnvironmentRepository`, `sessionQuerier` (interface) |
| `ProfileService` | `app/profile.go` | `ProfileRepository` |

## Domain Types Reference

| Type | File | Purpose |
|------|------|---------|
| `Session` | `domain/session.go` | Enriched view of a coding session |
| `Server` | `domain/session.go` | Summary of a discovered coding-agent server |
| `ManagedServer` | `domain/session.go` | Persistent state of an Orbit-launched server |
| `Environment` | `domain/environment.go` | Registered environment (path + metadata) |
| `Profile` | `domain/profile.go` | Reusable starter kit for environments |
| `CheckStatus` | `domain/doctor.go` | Doctor check result status (pass/warn/fail) |
| `CheckResult` | `domain/doctor.go` | Single diagnostic check outcome |
| `Report` | `domain/doctor.go` | Full set of doctor check results |
| `OpenAction` | `domain/open.go` | What the driver should do after resolving an env |
| `OpenPlan` | `domain/open.go` | Result of resolving an `orbit open` request |

## Drivers

### CLI (`internal/driver/cli/`)

Built with [cobra](https://github.com/spf13/cobra). The composition root
(`root.go`) wires adapters into app services. Individual commands receive
dependencies via function parameters or closures — they never import
adapter packages.

### HTTP Server (`internal/driver/server/`)

A standard `net/http` server. Defines its own service interfaces locally
so it has no import dependency on `internal/app`. Serves:

- REST API at `/api/*`
- Reverse proxy to OpenCode servers at `/api/proxy/{port}/{path...}`
- Embedded web UI (`index.html`) at `/`

### TUI (`internal/driver/tui/`)

Not yet implemented. Will use [bubbletea](https://github.com/charmbracelet/bubbletea)
and share the same app services as the CLI and HTTP drivers.

## Platform Considerations

Platform-specific code uses Go build tags:

| Function | Unix File | Windows File |
|----------|-----------|-------------|
| `interrupt()` | `process_signal.go` | `process_signal_windows.go` |
| `processAlive()` | `process_alive.go` | `process_alive_windows.go` |

The primary target is macOS/Linux. Windows support is best-effort.

## Conventions

- **No JSON tags on domain types.** Each adapter/driver owns its own DTO
  types with appropriate serialization tags.
- **Consumer-defined interfaces.** When a service or driver needs a subset
  of another component's methods, it defines a local interface rather than
  depending on the concrete type. This follows standard Go practice.
- **Atomic file writes.** Both `jsonstore` and `opencode/state.go` write
  to a temp file then rename, preventing corruption on crash.
- **Never delete files.** Archive to `~/Archive/` with timestamps instead.
- **`make check` is the pre-commit gate.** Runs `go fmt`, `go vet`, and
  `go test`.
