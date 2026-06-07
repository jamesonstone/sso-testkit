package redact

import (
	"net/http"
	"strings"
	"testing"
)

func TestStringRedactsKnownSecretsAndJWTs(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxMjMifQ.signature"
	secret := "super-secret"
	r := New(secret)
	out := r.String("token=" + token + " secret=" + secret)
	if strings.Contains(out, token) || strings.Contains(out, secret) {
		t.Fatalf("sensitive value leaked: %s", out)
	}
	if !strings.Contains(out, "[REDACTED-JWT:") || !strings.Contains(out, "[REDACTED:") {
		t.Fatalf("redaction markers missing: %s", out)
	}
}

func TestHeadersRedactSensitiveNames(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer secret-token")
	headers.Set("X-API-Key", "partner-api-key")
	headers.Set("x-request-id", "request-1")

	got := New().Headers(headers)
	if strings.Contains(got["Authorization"], "secret-token") {
		t.Fatalf("authorization leaked: %v", got)
	}
	if strings.Contains(got["X-API-Key"], "partner-api-key") {
		t.Fatalf("api key leaked: %v", got)
	}
	if got["X-Request-Id"] != "request-1" {
		t.Fatalf("non-sensitive header changed: %v", got)
	}
}
