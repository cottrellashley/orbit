package domain_test

import (
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestCheckStatusString(t *testing.T) {
	tests := []struct {
		status domain.CheckStatus
		want   string
	}{
		{domain.CheckPass, "pass"},
		{domain.CheckWarn, "warn"},
		{domain.CheckFail, "fail"},
		{domain.CheckStatus(99), "unknown"},
	}
	for _, tc := range tests {
		got := tc.status.String()
		if got != tc.want {
			t.Errorf("CheckStatus(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestReportOK(t *testing.T) {
	t.Run("empty report is OK", func(t *testing.T) {
		r := &domain.Report{}
		if !r.OK() {
			t.Error("expected empty report to be OK")
		}
	})

	t.Run("all pass is OK", func(t *testing.T) {
		r := &domain.Report{Results: []domain.CheckResult{
			{Status: domain.CheckPass},
			{Status: domain.CheckPass},
		}}
		if !r.OK() {
			t.Error("expected all-pass report to be OK")
		}
	})

	t.Run("warn makes report not OK", func(t *testing.T) {
		r := &domain.Report{Results: []domain.CheckResult{
			{Status: domain.CheckPass},
			{Status: domain.CheckWarn},
		}}
		if r.OK() {
			t.Error("expected report with warning to not be OK")
		}
	})

	t.Run("fail makes report not OK", func(t *testing.T) {
		r := &domain.Report{Results: []domain.CheckResult{
			{Status: domain.CheckFail},
		}}
		if r.OK() {
			t.Error("expected report with failure to not be OK")
		}
	})
}

func TestReportFailures(t *testing.T) {
	r := &domain.Report{Results: []domain.CheckResult{
		{Name: "a", Status: domain.CheckPass},
		{Name: "b", Status: domain.CheckFail},
		{Name: "c", Status: domain.CheckWarn},
		{Name: "d", Status: domain.CheckFail},
	}}
	failures := r.Failures()
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures, got %d", len(failures))
	}
	if failures[0].Name != "b" || failures[1].Name != "d" {
		t.Errorf("unexpected failure names: %v", failures)
	}
}

func TestReportWarnings(t *testing.T) {
	r := &domain.Report{Results: []domain.CheckResult{
		{Name: "a", Status: domain.CheckPass},
		{Name: "b", Status: domain.CheckWarn},
		{Name: "c", Status: domain.CheckFail},
		{Name: "d", Status: domain.CheckWarn},
	}}
	warnings := r.Warnings()
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(warnings))
	}
	if warnings[0].Name != "b" || warnings[1].Name != "d" {
		t.Errorf("unexpected warning names: %v", warnings)
	}
}

func TestReportFailuresEmpty(t *testing.T) {
	r := &domain.Report{Results: []domain.CheckResult{
		{Status: domain.CheckPass},
		{Status: domain.CheckWarn},
	}}
	if got := r.Failures(); got != nil {
		t.Errorf("expected nil failures slice, got %v", got)
	}
}

func TestReportWarningsEmpty(t *testing.T) {
	r := &domain.Report{Results: []domain.CheckResult{
		{Status: domain.CheckPass},
		{Status: domain.CheckFail},
	}}
	if got := r.Warnings(); got != nil {
		t.Errorf("expected nil warnings slice, got %v", got)
	}
}
