package config

import (
	"strings"
	"testing"
)

const validConfig = `
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
  extra_values: []
`

func TestParseAndValidateStub(t *testing.T) {
	cfg, err := Parse([]byte(validConfig))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate("", func(string) (string, bool) { return "", false }); err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != ModeStub {
		t.Fatalf("Mode = %q", cfg.Mode)
	}
}

func TestRejectsUnknownFields(t *testing.T) {
	_, err := Parse([]byte(validConfig + "\nunknown: true\n"))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestRejectsLiteralSensitiveHeaders(t *testing.T) {
	raw := strings.Replace(validConfig, "X-API-Key: env:PARTNER_API_KEY", "X-API-Key: literal-key", 1)
	cfg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	err = cfg.Validate("", func(string) (string, bool) { return "", false })
	if err == nil || !strings.Contains(err.Error(), "exchange.headers.X-API-Key must use env:VARIABLE_NAME") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLiveModeResolvesSecrets(t *testing.T) {
	cfg, err := Parse([]byte(validConfig))
	if err != nil {
		t.Fatal(err)
	}
	err = cfg.Validate(ModeLive, func(name string) (string, bool) {
		if name == "PARTNER_API_KEY" {
			return "secret-client-key", true
		}
		return "", false
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.ResolvedSecrets["exchange.headers.X-API-Key"]; got != "secret-client-key" {
		t.Fatalf("resolved secret = %q", got)
	}
}

func TestLiveModeFailsUnsetEnv(t *testing.T) {
	cfg, err := Parse([]byte(validConfig))
	if err != nil {
		t.Fatal(err)
	}
	err = cfg.Validate(ModeLive, func(string) (string, bool) { return "", false })
	if err == nil || !strings.Contains(err.Error(), "PARTNER_API_KEY") {
		t.Fatalf("expected missing env error, got %v", err)
	}
}

func TestRejectsInvalidCallbackConfig(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "missing redirect uri",
			raw:  strings.Replace(validConfig, "  redirect_uri: http://127.0.0.1:8787/callback\n", "", 1),
			want: "callback.redirect_uri or idp.redirect_uri is required",
		},
		{
			name: "path without slash",
			raw:  strings.Replace(validConfig, "  path: /callback", "  path: callback", 1),
			want: "callback.path must start with /",
		},
		{
			name: "port out of range",
			raw:  strings.Replace(validConfig, "  port: 8787", "  port: 70000", 1),
			want: "callback.port must be between 0 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse([]byte(tt.raw))
			if err != nil {
				t.Fatal(err)
			}
			err = cfg.Validate("", func(string) (string, bool) { return "", false })
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}
