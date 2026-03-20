// Package port defines the interfaces (ports) that the domain and application
// layers depend on. Adapters implement these interfaces.
//
// Secondary (driven) ports:
//   - EnvironmentRepository: persist and retrieve environment registry entries
//   - ProjectRepository: persist and retrieve project registry entries (migration successor to EnvironmentRepository)
//   - ProfileRepository: manage profile starter kits on disk
//   - SessionProvider: discover servers, list/manage coding sessions
//   - ServerLifecycle: start, stop, and monitor a managed coding-agent server
//   - WorkspaceManager: orchestrate terminal multiplexer windows
//   - ConfigWorkspace: resolve Orbit config directory layout and policies
//   - GitHubProvider: authenticate and query the GitHub REST API (repos, issues, links)
//   - MarkdownRenderer: convert markdown text to HTML
//   - ToolInstaller: check installation status and install developer tools
//   - CopilotTaskProvider: discover and manage Copilot coding agent tasks
//
// Primary (driving) ports are the app service methods themselves — drivers
// (CLI, TUI, HTTP server) call them directly.
package port
