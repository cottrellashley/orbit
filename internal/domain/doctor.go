package domain

// CheckStatus represents the result status of a single doctor check.
type CheckStatus int

const (
	CheckPass CheckStatus = iota
	CheckWarn
	CheckFail
)

// String returns a human-readable label for the status.
func (s CheckStatus) String() string {
	switch s {
	case CheckPass:
		return "pass"
	case CheckWarn:
		return "warn"
	case CheckFail:
		return "fail"
	default:
		return "unknown"
	}
}

// CheckResult is the outcome of a single diagnostic check.
type CheckResult struct {
	Name    string      // short identifier, e.g. "opencode"
	Status  CheckStatus // pass, warn, or fail
	Message string      // human-readable description of the finding
	Fix     string      // actionable fix hint (empty when status is pass)
}

// Report is the full set of check results from a doctor run.
type Report struct {
	Results []CheckResult
}

// OK returns true if every check passed (no warnings or failures).
func (r *Report) OK() bool {
	for _, c := range r.Results {
		if c.Status != CheckPass {
			return false
		}
	}
	return true
}

// Failures returns only the checks that failed.
func (r *Report) Failures() []CheckResult {
	var out []CheckResult
	for _, c := range r.Results {
		if c.Status == CheckFail {
			out = append(out, c)
		}
	}
	return out
}

// Warnings returns only the checks that produced warnings.
func (r *Report) Warnings() []CheckResult {
	var out []CheckResult
	for _, c := range r.Results {
		if c.Status == CheckWarn {
			out = append(out, c)
		}
	}
	return out
}
