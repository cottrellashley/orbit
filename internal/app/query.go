package app

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// StateQuerier provides a read-only facade over Orbit's runtime state.
// It is designed for consumers that need a unified view without mutating
// anything — in particular, the future Orbit Assistant driver.
//
// All methods are failure-tolerant: if a subsystem is unavailable the
// result is empty/nil rather than an error, unless the error is truly
// unexpected. This supports partial-availability scenarios where e.g.
// session discovery fails but project data is still readable.
type StateQuerier struct {
	projects *ProjectService
	sessions *SessionService
	doctor   *DoctorService
	open     *OpenService
}

// NewStateQuerier creates a StateQuerier.
// Any parameter may be nil; the corresponding methods will return
// zero-value results.
func NewStateQuerier(
	projects *ProjectService,
	sessions *SessionService,
	doctor *DoctorService,
	open *OpenService,
) *StateQuerier {
	return &StateQuerier{
		projects: projects,
		sessions: sessions,
		doctor:   doctor,
		open:     open,
	}
}

// Projects returns all registered projects. Returns nil on error.
func (q *StateQuerier) Projects() []*domain.Project {
	if q.projects == nil {
		return nil
	}
	list, err := q.projects.List()
	if err != nil {
		return nil
	}
	return list
}

// Project returns a single project by name. Returns nil if not found.
func (q *StateQuerier) Project(name string) *domain.Project {
	if q.projects == nil {
		return nil
	}
	p, err := q.projects.Get(name)
	if err != nil {
		return nil
	}
	return p
}

// Sessions returns all sessions across all discovered servers.
// Returns nil on error.
func (q *StateQuerier) Sessions(ctx context.Context) []domain.Session {
	if q.sessions == nil {
		return nil
	}
	list, err := q.sessions.ListAll(ctx)
	if err != nil {
		return nil
	}
	return list
}

// Servers returns all discovered coding-agent servers.
// Returns nil on error.
func (q *StateQuerier) Servers(ctx context.Context) []domain.Server {
	if q.sessions == nil {
		return nil
	}
	list, err := q.sessions.DiscoverServers(ctx)
	if err != nil {
		return nil
	}
	return list
}

// DoctorReport runs all diagnostic checks and returns the report.
// Returns nil if the doctor service is not configured.
func (q *StateQuerier) DoctorReport(ctx context.Context) *domain.Report {
	if q.doctor == nil {
		return nil
	}
	return q.doctor.Run(ctx)
}

// ProjectPlan resolves the open plan for a project by name.
// Returns nil on error or if project support is not configured.
func (q *StateQuerier) ProjectPlan(ctx context.Context, projectName string) *domain.ProjectOpenPlan {
	if q.open == nil {
		return nil
	}
	plan, err := q.open.ResolveProject(ctx, projectName)
	if err != nil {
		return nil
	}
	return plan
}
