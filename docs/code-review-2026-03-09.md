# Code Review: Hexagonal Architecture Audit

**Date:** 2026-03-09
**Scope:** Full codebase review of `internal/` for hexagonal architecture violations.

## Summary

10 issues identified across 4 severity levels. All 10 have been fixed.
`make check` passes clean after all changes.

---

## Critical Findings (4)

### 1. SessionService bypassed lifecycle — runtime bug

**Files:** `internal/app/session.go`

`ListAll()`, `GetSession()`, `CreateSession()`, and `findServerForSession()`
all called `s.provider.DiscoverServers(ctx)` directly instead of
`s.DiscoverServers(ctx)`. The public `DiscoverServers()` method merges the
managed server (from `ServerLifecycle`) with process-table discovery results.
The private calls skipped this merge entirely.

**Impact:** The managed OpenCode server launched by `orbit serve` was
invisible to all session operations. Sessions on the managed server could
not be listed, fetched, created, aborted, or deleted.

**Fix:** Changed all four call sites to use `s.DiscoverServers(ctx)`.

### 2. `attach.go` imported adapter directly

**File:** `internal/driver/cli/attach.go`

The `orbit attach` command imported
`oc "github.com/cottrellashley/orbit/internal/adapter/opencode"` and
instantiated `oc.NewServerManager(oc.ServerManagerOpts{})` inline.

**Violation:** Drivers must never reach into adapters. They go through
app services or receive port interfaces from the composition root.

**Fix:** `newAttachCmd()` now accepts a `port.ServerLifecycle` parameter.
The composition root in `root.go` creates the `ServerManager` and passes
it in.

### 3. `root.go` leaked adapter types into command logic

**File:** `internal/driver/cli/root.go`

While `root.go` is correctly the composition root (allowed to import
adapters), the `newServeCmd()` function directly used `oc.NewServerManager`,
`oc.ServerManagerOpts`, `oc.ProcessOpts`, and `oc.DefaultStopTimeout`.
This mixed DI wiring with command business logic.

**Fix:** Extracted `newServerManager(dir string) port.ServerLifecycle`
factory function. Introduced local `defaultStopTimeout` constant.
`newServeCmd()` now uses only `port.ServerLifecycle` — no adapter types.

### 4. `OpenService` depended on concrete `*SessionService`

**File:** `internal/app/open.go`

`OpenService.sessions` was typed as `*SessionService` (concrete struct).
App services should depend on interfaces, not on each other's concrete
types.

**Fix:** Introduced `sessionQuerier` interface (consumer-defined) with
`ListForEnvironment()` and `DiscoverServers()`. `*SessionService`
satisfies this interface implicitly.

---

## Moderate Findings (4)

### 5. HTTP server depended on concrete app service types

**File:** `internal/driver/server/server.go`

All constructor parameters were `*app.EnvironmentService`,
`*app.SessionService`, etc., creating a direct import dependency on
`internal/app`.

**Fix:** Defined local interfaces (`EnvironmentService`, `SessionService`,
`DoctorService`, `OpenService`) in the server package. The `internal/app`
import is gone.

### 6. Domain types carried JSON struct tags

**Files:** `internal/domain/session.go`, `internal/domain/environment.go`

`ManagedServer` and `Environment` had `json:"..."` struct tags.
Serialization format is an adapter/infrastructure concern, not a domain
concern.

**Fix:** Removed all JSON tags from domain types. Added DTO types:
- `jsonstore.envDTO` with JSON tags, plus `toDTO`/`fromDTO` converters.
- `opencode.managedServerDTO` with JSON tags, plus
  `toManagedDTO`/`fromManagedDTO` converters.
- API response DTOs (`serverJSON`, `sessionJSON`, `environmentJSON`) in
  the HTTP server driver.

### 7. Doctor types mixed with Open types in `domain/open.go`

**File:** `internal/domain/open.go`

`CheckStatus`, `CheckResult`, and `Report` (doctor concepts) were in the
same file as `OpenAction` and `OpenPlan`. Misleading file naming and
unrelated concepts grouped together.

**Fix:** Moved doctor types to new file `internal/domain/doctor.go`.
`open.go` now contains only `OpenAction` and `OpenPlan`.

### 8. `processAlive()` used `syscall.Signal(0)` without build tags

**File:** `internal/adapter/opencode/state.go`

`processAlive()` called `proc.Signal(syscall.Signal(0))` which is
Unix-specific. The build-tag pattern already existed for `interrupt()`
(`process_signal.go` / `process_signal_windows.go`) but was not applied
here.

**Fix:** Extracted `processAlive()` from `state.go` into:
- `process_alive.go` (`//go:build !windows`) — uses `syscall.Signal(0)`
- `process_alive_windows.go` (`//go:build windows`) — uses `tasklist`

Removed `syscall` import from `state.go`.

---

## Minor Findings (2)

### 9. Inconsistent JSON tag policy across domain types

`domain.Server` had no JSON tags while `domain.ManagedServer` did.
Resolved as part of finding #6 — all domain types are now tag-free.

### 10. `ServerLifecycle.Server()` lacked `context.Context`

**File:** `internal/port/server_lifecycle.go`

The `Server()` method performed I/O (reads state file, checks process
liveness) but took no context parameter.

**Fix:** Changed signature to `Server(ctx context.Context) *domain.Server`.
Updated implementation in `lifecycle.go` and caller in
`SessionService.DiscoverServers()`.

---

## Files Changed

| File | Change |
|------|--------|
| `internal/app/session.go` | Fixed 4 call sites to use `s.DiscoverServers(ctx)` |
| `internal/app/open.go` | `sessionQuerier` interface replaces concrete dependency |
| `internal/driver/cli/attach.go` | Receives `port.ServerLifecycle` parameter |
| `internal/driver/cli/root.go` | `newServerManager()` factory, local timeout constant |
| `internal/driver/server/server.go` | Local service interfaces, API response DTOs |
| `internal/domain/session.go` | Removed JSON tags from `ManagedServer` |
| `internal/domain/environment.go` | Removed JSON tags from `Environment` |
| `internal/domain/open.go` | Removed doctor types |
| `internal/domain/doctor.go` | **New** — `CheckStatus`, `CheckResult`, `Report` |
| `internal/port/server_lifecycle.go` | Added `context.Context` to `Server()` |
| `internal/adapter/opencode/state.go` | `managedServerDTO`, removed `processAlive` |
| `internal/adapter/opencode/lifecycle.go` | Updated `Server()` signature |
| `internal/adapter/opencode/process_alive.go` | **New** — Unix `processAlive()` |
| `internal/adapter/opencode/process_alive_windows.go` | **New** — Windows `processAlive()` |
| `internal/adapter/jsonstore/store.go` | `envDTO` with `toDTO`/`fromDTO` converters |

---

## Verification

```
$ make check
go fmt ./...
go vet ./...
go test ./... -v
```

All pass clean. No test files exist yet (tracked separately).
