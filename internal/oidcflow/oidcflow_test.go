package oidcflow

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
	"github.com/jamesonstone/sso-testkit/internal/testsupport"
)

func TestVerifyIDToken(t *testing.T) {
	server := testsupport.NewOIDCServer(t)
	ctx := oidc.ClientContext(context.Background(), server.Server.Client())
	provider, err := oidc.NewProvider(ctx, server.Issuer)
	if err != nil {
		t.Fatal(err)
	}
	raw := server.IDToken(t, "nonce", time.Now().Add(time.Hour), nil)

	claims, check := VerifyIDToken(ctx, provider, raw, server.ClientID, "nonce")
	if check.Status != report.StatusPass {
		t.Fatalf("check = %#v", check)
	}
	if claims["sub"] != "user-123" {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestVerifyIDTokenRejectsNonceMismatch(t *testing.T) {
	server := testsupport.NewOIDCServer(t)
	ctx := oidc.ClientContext(context.Background(), server.Server.Client())
	provider, err := oidc.NewProvider(ctx, server.Issuer)
	if err != nil {
		t.Fatal(err)
	}
	raw := server.IDToken(t, "nonce", time.Now().Add(time.Hour), nil)

	_, check := VerifyIDToken(ctx, provider, raw, server.ClientID, "other")
	if check.Status != report.StatusFail || check.Classification != "nonce_mismatch" {
		t.Fatalf("check = %#v", check)
	}
}

func TestVerifyIDTokenRejectsInvalidSignature(t *testing.T) {
	providerServer := testsupport.NewOIDCServer(t)
	tokenServer := testsupport.NewOIDCServer(t)
	ctx := oidc.ClientContext(context.Background(), providerServer.Server.Client())
	provider, err := oidc.NewProvider(ctx, providerServer.Issuer)
	if err != nil {
		t.Fatal(err)
	}
	raw := tokenServer.IDToken(t, "nonce", time.Now().Add(time.Hour), map[string]any{
		"iss": providerServer.Issuer,
		"aud": providerServer.ClientID,
	})

	_, check := VerifyIDToken(ctx, provider, raw, providerServer.ClientID, "nonce")
	if check.Status != report.StatusFail || check.Classification != "id_token_validation" {
		t.Fatalf("check = %#v", check)
	}
}

func TestVerifyIDTokenRejectsIssuerAudienceAndExpiry(t *testing.T) {
	server := testsupport.NewOIDCServer(t)
	ctx := oidc.ClientContext(context.Background(), server.Server.Client())
	provider, err := oidc.NewProvider(ctx, server.Issuer)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		token     string
		clientID  string
		nonce     string
		wantClass string
	}{
		{
			name:     "issuer",
			token:    server.IDToken(t, "nonce", time.Now().Add(time.Hour), map[string]any{"iss": "https://issuer.example.invalid"}),
			clientID: server.ClientID,
			nonce:    "nonce",
		},
		{
			name:     "audience",
			token:    server.IDToken(t, "nonce", time.Now().Add(time.Hour), nil),
			clientID: "other-client",
			nonce:    "nonce",
		},
		{
			name:     "expired",
			token:    server.IDToken(t, "nonce", time.Now().Add(-time.Hour), nil),
			clientID: server.ClientID,
			nonce:    "nonce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, check := VerifyIDToken(ctx, provider, tt.token, tt.clientID, tt.nonce)
			if check.Status != report.StatusFail {
				t.Fatalf("check = %#v", check)
			}
		})
	}
}

func TestValidateClaims(t *testing.T) {
	check := ValidateClaims(map[string]any{"sub": "user-123"}, []config.ClaimRule{{Name: "sub", Required: true, ExpectedValue: "user-123"}})
	if check.Status != report.StatusPass {
		t.Fatalf("check = %#v", check)
	}
	missing := ValidateClaims(map[string]any{}, []config.ClaimRule{{Name: "sub", Required: true}})
	if missing.Status != report.StatusFail || !strings.Contains(missing.FailureReason, "sub") {
		t.Fatalf("missing = %#v", missing)
	}
}

func TestWaitForCallbackRejectsStateMismatchAndMissingCode(t *testing.T) {
	tests := []struct {
		name  string
		query string
		class string
	}{
		{name: "state", query: "state=wrong&code=abc", class: "state_mismatch"},
		{name: "code", query: "state=expected", class: "missing_code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := freePort(t)
			done := make(chan report.Check, 1)
			go func() {
				_, check := waitForCallback(context.Background(), config.Callback{
					Bind:    "127.0.0.1",
					Port:    port,
					Path:    "/callback",
					Timeout: "2s",
				}, "expected")
				done <- check
			}()

			url := "http://127.0.0.1:" + fmt.Sprint(port) + "/callback?" + tt.query
			var lastErr error
			for i := 0; i < 20; i++ {
				resp, err := http.Get(url)
				if err == nil {
					resp.Body.Close()
					break
				}
				lastErr = err
				time.Sleep(10 * time.Millisecond)
			}
			select {
			case check := <-done:
				if check.Status != report.StatusFail || check.Classification != tt.class {
					t.Fatalf("check = %#v", check)
				}
			case <-time.After(3 * time.Second):
				t.Fatalf("callback did not finish, last HTTP error: %v", lastErr)
			}
		})
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func TestRunStub(t *testing.T) {
	result := RunStub(config.Config{
		IDP:            config.IDP{Issuer: "https://issuer.example", ClientID: "client"},
		ExpectedClaims: []config.ClaimRule{{Name: "sub", Required: true}},
	}, redact.New())
	if result.IDToken == "" {
		t.Fatal("stub ID token missing")
	}
	if result.Checks[len(result.Checks)-1].Status != report.StatusPass {
		t.Fatalf("checks = %#v", result.Checks)
	}
}
