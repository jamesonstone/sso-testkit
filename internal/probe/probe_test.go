package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
)

func TestRunProbeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	checks := Run(context.Background(), config.Probe{Enabled: true, URL: server.URL, ExpectedStatus: []int{204}}, nil, "token", server.Client(), redact.New("token"))
	if checks[len(checks)-1].Status != report.StatusPass {
		t.Fatalf("checks = %#v", checks)
	}
}

func TestRunProbeUsesResolvedHeaderSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Api-Key"); got != "resolved-key" {
			t.Fatalf("X-Api-Key = %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.Probe{
		Enabled: true,
		URL:     server.URL,
		Headers: map[string]string{
			"X-Api-Key": "env:PROBE_API_KEY",
		},
	}
	checks := Run(context.Background(), cfg, map[string]string{"probe.headers.X-Api-Key": "resolved-key"}, "", server.Client(), redact.New("resolved-key"))
	if checks[len(checks)-1].Status != report.StatusPass {
		t.Fatalf("checks = %#v", checks)
	}
}

func TestRunProbeClassifiesFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "do not include token", http.StatusTooManyRequests)
	}))
	defer server.Close()

	checks := Run(context.Background(), config.Probe{Enabled: true, URL: server.URL, RedactBody: true}, nil, "token", server.Client(), redact.New("token"))
	last := checks[len(checks)-1]
	if last.Status != report.StatusFail || last.Classification != "rate_limited" {
		t.Fatalf("last check = %#v", last)
	}
	if last.Evidence["body_summary"] == "do not include token\n" {
		t.Fatalf("expected body to pass through redactor: %#v", last)
	}
}
