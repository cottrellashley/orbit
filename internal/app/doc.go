// Package app contains Orbit's application services.
//
// Application services orchestrate domain logic and port interactions.
// They are the primary ports — drivers (CLI, TUI, HTTP) call these
// methods directly. Services never depend on concrete adapters, only
// on port interfaces.
package app
