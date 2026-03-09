// Package opencode implements port.SessionProvider by wrapping the official
// OpenCode Go SDK (github.com/sst/opencode-sdk-go).
//
// The package has two concerns:
//
//   - Adapter: translates between SDK types (opencode.Session, opencode.Path)
//     and Orbit domain types (domain.Session, domain.Server), implementing
//     the SessionProvider interface that the app layer depends on.
//   - Process: manages OpenCode server lifecycle — starting servers, discovering
//     running instances via the process table, and health-checking candidates.
//     Process management is not covered by the SDK and uses os/exec directly.
//
// Endpoints not yet available in the SDK (health, dispose, session status)
// are accessed through the SDK client's raw Get/Post escape-hatch methods.
package opencode
