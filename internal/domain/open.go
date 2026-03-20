package domain

import "fmt"

// OpenAction describes what the CLI/TUI should do after resolving an environment.
type OpenAction int

const (
	// OpenActionCreate indicates no sessions exist; driver should offer to create one.
	OpenActionCreate OpenAction = iota
	// OpenActionResume indicates exactly one session exists; driver should offer
	// to resume it or create a new one.
	OpenActionResume
	// OpenActionSelect indicates multiple sessions exist; driver should prompt
	// the user to pick one or create a new one.
	OpenActionSelect
)

// String returns a human-readable label.
func (a OpenAction) String() string {
	switch a {
	case OpenActionCreate:
		return "create"
	case OpenActionResume:
		return "resume"
	case OpenActionSelect:
		return "select"
	default:
		return fmt.Sprintf("OpenAction(%d)", int(a))
	}
}

// OpenPlan is the result of resolving an `orbit open` request. It contains
// everything the driver layer needs to present options to the user.
type OpenPlan struct {
	Environment *Environment // the resolved environment
	Sessions    []Session    // active sessions for this environment (may be empty)
	Action      OpenAction   // what the driver should do
}

// ProjectOpenPlan is the Project-first equivalent of OpenPlan. It carries
// a *Project instead of an *Environment and adds server presence info.
// During the migration period both plan types may coexist; drivers choose
// which to consume based on whether they've migrated to the project model.
type ProjectOpenPlan struct {
	Project      *Project   // the resolved project
	Sessions     []Session  // active sessions for this project (may be empty)
	Action       OpenAction // what the driver should do
	ServerOnline bool       // true if a coding-agent server is running for this project
}
