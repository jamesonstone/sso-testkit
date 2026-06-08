package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	path := writeScenario(t)
	var stdout, stderr bytes.Buffer
	code := Main([]string{"validate-config", "--config", path}, &stdout, &stderr, func(string) (string, bool) { return "", false })
	if code != ExitOK {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "config ok: oidc-token-exchange mode=stub") {
		t.Fatalf("stdout=%s", stdout.String())
	}
}

func TestRunStubWritesReport(t *testing.T) {
	path := writeScenario(t)
	var stdout, stderr bytes.Buffer
	code := Main([]string{"run", "--config", path, "--mode", "stub", "--report", "-"}, &stdout, &stderr, func(string) (string, bool) { return "", false })
	if code != ExitOK {
		t.Fatalf("code=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), `"overall_status": "pass"`) {
		t.Fatalf("report missing pass status: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "stub-token") {
		t.Fatalf("stub token leaked in report: %s", stdout.String())
	}
}

func TestBadArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Main([]string{"run"}, &stdout, &stderr, os.LookupEnv)
	if code != ExitUsage {
		t.Fatalf("code=%d", code)
	}
}

func writeScenario(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(testScenario), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const testScenario = `
scenario:
  id: oidc-token-exchange
  name: OIDC token exchange
  description: test scenario
  providers:
    idp: example-idp
    sp: example-api
mode_defaults:
  mode: stub
callback:
  bind: 127.0.0.1
  port: 8787
  path: /callback
  redirect_uri: http://127.0.0.1:8787/callback
  timeout: 5m
idp:
  issuer: https://issuer.example.com/
  client_id: test-client
  scopes:
    - openid
    - profile
  expected_algs:
    - RS256
expected_claims:
  - name: sub
    required: true
exchange:
  profile: rfc8693
  endpoint: https://service-provider.example.com/oauth/token
  method: POST
  headers:
    X-API-Key: env:PARTNER_API_KEY
  subject_token_source: id_token
  subject_token_type: urn:ietf:params:oauth:token-type:jwt
  requested_token_type: urn:ietf:params:oauth:token-type:access_token
  timeout: 10s
  stub_response:
    access_token: stub-token
    issued_token_type: urn:ietf:params:oauth:token-type:access_token
    token_type: Bearer
    expires_in: 3600
probe:
  enabled: false
  expected_status:
    - 200
  timeout: 10s
  redact_body: true
report:
  output: json
redaction:
  extra_values:
    - stub-token
`
