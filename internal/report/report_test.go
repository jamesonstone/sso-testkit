package report

import (
	"strings"
	"testing"
	"time"
)

func TestOverallStatus(t *testing.T) {
	tests := []struct {
		name   string
		checks []Check
		want   Status
	}{
		{name: "empty", want: StatusSkip},
		{name: "all pass", checks: []Check{{Status: StatusPass}}, want: StatusPass},
		{name: "all skip", checks: []Check{{Status: StatusSkip}}, want: StatusSkip},
		{name: "blocked", checks: []Check{{Status: StatusPass}, {Status: StatusBlocked}}, want: StatusBlocked},
		{name: "fail wins", checks: []Check{{Status: StatusBlocked}, {Status: StatusFail}}, want: StatusFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OverallStatus(tt.checks); got != tt.want {
				t.Fatalf("OverallStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReportJSONDeterministic(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	r := New("scenario", "stub", now)
	r.Add(Pass(CheckConfigLoad, map[string]string{"path": "config.yaml"}))
	r.Add(Blocked(CheckExchangeResp, "external_trust", "trust is incomplete", "external_provider", "confirm trust and client configuration", nil))
	r.Finish(now.Add(time.Second))

	first, err := r.JSON()
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("JSON output changed between renders\nfirst=%s\nsecond=%s", first, second)
	}
	if !strings.Contains(string(first), `"overall_status": "blocked"`) {
		t.Fatalf("expected blocked overall status in %s", first)
	}
}
