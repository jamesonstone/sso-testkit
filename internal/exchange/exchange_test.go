package exchange

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
)

func TestBuildRequestIncludesRFC8693FieldsAndResolvedHeader(t *testing.T) {
	cfg := config.Exchange{
		Endpoint:           "https://example.com/token",
		Method:             http.MethodPost,
		Headers:            map[string]string{"X-API-Key": "env:PARTNER_API_KEY"},
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
	}
	req, err := BuildRequest(context.Background(), cfg, map[string]string{"exchange.headers.X-API-Key": "secret-key"}, "subject-token")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	form := string(body)
	for _, want := range []string{
		"grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Atoken-exchange",
		"subject_token=subject-token",
		"subject_token_type=urn%3Aietf%3Aparams%3Aoauth%3Atoken-type%3Ajwt",
		"requested_token_type=urn%3Aietf%3Aparams%3Aoauth%3Atoken-type%3Aaccess_token",
	} {
		if !strings.Contains(form, want) {
			t.Fatalf("form %q missing %q", form, want)
		}
	}
	if got := req.Header.Get("X-API-Key"); got != "secret-key" {
		t.Fatalf("X-API-Key = %q", got)
	}
}

func TestRunRedactsTokens(t *testing.T) {
	const subject = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxIn0.signature"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.Form.Get("subject_token"); got != subject {
			t.Fatalf("subject_token = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"returned-token","issued_token_type":"urn:ietf:params:oauth:token-type:access_token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	cfg := config.Exchange{
		Endpoint:           server.URL,
		Method:             http.MethodPost,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
	}
	token, checks := Run(context.Background(), cfg, nil, subject, server.Client(), redact.New(subject, "returned-token"), config.ModeLive)
	if token.AccessToken != "returned-token" {
		t.Fatalf("AccessToken = %q", token.AccessToken)
	}
	for _, check := range checks {
		if strings.Contains(check.FailureReason, subject) || strings.Contains(check.FailureReason, "returned-token") {
			t.Fatalf("secret leaked in check: %#v", check)
		}
	}
}

func TestRunClassifiesExternalTrustFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "trust not configured", http.StatusForbidden)
	}))
	defer server.Close()

	cfg := config.Exchange{
		Profile:            "rfc8693",
		Endpoint:           server.URL,
		Method:             http.MethodPost,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
	}
	_, checks := Run(context.Background(), cfg, nil, "subject", server.Client(), redact.New("subject"), config.ModeLive)
	last := checks[len(checks)-1]
	if last.Status != report.StatusBlocked || last.Classification != "external_trust" {
		t.Fatalf("last check = %#v", last)
	}
}
