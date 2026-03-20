package domain_test

import (
	"fmt"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestOpenActionString(t *testing.T) {
	tests := []struct {
		action domain.OpenAction
		want   string
	}{
		{domain.OpenActionCreate, "create"},
		{domain.OpenActionResume, "resume"},
		{domain.OpenActionSelect, "select"},
		{domain.OpenAction(99), fmt.Sprintf("OpenAction(%d)", 99)},
	}
	for _, tc := range tests {
		got := tc.action.String()
		if got != tc.want {
			t.Errorf("OpenAction(%d).String() = %q, want %q", tc.action, got, tc.want)
		}
	}
}
