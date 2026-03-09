// Package port defines the interfaces (ports) that the domain and application
// layers depend on. Adapters implement these interfaces.
//
// Secondary (driven) ports:
//   - EnvironmentRepository: persist and retrieve environment registry entries
//   - ProfileRepository: manage profile starter kits on disk
//   - SessionProvider: discover servers, list/manage coding sessions
//   - ServerLifecycle: start, stop, and monitor a managed coding-agent server
//   - WorkspaceManager: orchestrate terminal multiplexer windows
//
// Primary (driving) ports are the app service methods themselves — drivers
// (CLI, TUI, HTTP server) call them directly.
package port
