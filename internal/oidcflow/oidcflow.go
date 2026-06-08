package oidcflow

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
	"golang.org/x/oauth2"
)

type Result struct {
	IDToken string
	Claims  map[string]any
	Checks  []report.Check
}

func RunStub(cfg config.Config, redactor redact.Redactor) Result {
	claims := map[string]any{"sub": "stub-user", "iss": cfg.IDP.Issuer, "aud": cfg.IDP.ClientID}
	checks := []report.Check{
		report.Pass(report.CheckOIDCDiscovery, map[string]string{"mode": "stub", "issuer": cfg.IDP.Issuer}),
		report.Pass(report.CheckOIDCAuthURL, map[string]string{"mode": "stub"}),
		report.Pass(report.CheckOIDCCallback, map[string]string{"mode": "stub"}),
		report.Pass(report.CheckOIDCToken, map[string]string{"mode": "stub"}),
		report.Pass(report.CheckOIDCIDToken, map[string]string{"token_fingerprint": redact.Fingerprint("stub-id-token")}),
	}
	claimCheck := ValidateClaims(claims, cfg.ExpectedClaims)
	if claimCheck.Status == report.StatusFail {
		checks = append(checks, claimCheck)
	} else {
		checks = append(checks, report.Pass(report.CheckOIDCClaims, map[string]string{"claims": redactor.String("sub,aud,iss")}))
	}
	return Result{IDToken: "stub-id-token", Claims: claims, Checks: checks}
}

func RunLive(ctx context.Context, cfg config.Config, client *http.Client, notifyURL func(string)) Result {
	redactor := redact.New(cfg.SecretValues()...)
	for _, value := range cfg.Redaction.ExtraValues {
		redactor.Add(value)
	}
	ctx = oidc.ClientContext(ctx, client)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)

	provider, err := oidc.NewProvider(ctx, cfg.IDP.Issuer)
	if err != nil {
		return Result{Checks: []report.Check{report.Fail(report.CheckOIDCDiscovery, "discovery", redactor.String(err.Error()), map[string]string{"issuer": cfg.IDP.Issuer})}}
	}
	checks := []report.Check{report.Pass(report.CheckOIDCDiscovery, map[string]string{"issuer": cfg.IDP.Issuer})}

	verifier, err := RandomURLSafe(64)
	if err != nil {
		return Result{Checks: append(checks, report.Fail(report.CheckOIDCAuthURL, "random", err.Error(), nil))}
	}
	state, err := RandomURLSafe(32)
	if err != nil {
		return Result{Checks: append(checks, report.Fail(report.CheckOIDCAuthURL, "random", err.Error(), nil))}
	}
	nonce, err := RandomURLSafe(32)
	if err != nil {
		return Result{Checks: append(checks, report.Fail(report.CheckOIDCAuthURL, "random", err.Error(), nil))}
	}

	oauthConfig := oauth2.Config{
		ClientID:     cfg.IDP.ClientID,
		ClientSecret: cfg.ResolvedSecrets["idp.client_secret"],
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURI(),
		Scopes:       cfg.IDP.Scopes,
	}
	authURL := oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("nonce", nonce))
	checks = append(checks, report.Pass(report.CheckOIDCAuthURL, map[string]string{"auth_url": redactor.URL(authURL)}))
	if notifyURL != nil {
		notifyURL(authURL)
	}

	callback, callbackCheck := waitForCallback(ctx, cfg.Callback, state)
	checks = append(checks, callbackCheck)
	if callbackCheck.Status != report.StatusPass {
		return Result{Checks: checks}
	}

	token, err := oauthConfig.Exchange(ctx, callback.Code, oauth2.VerifierOption(verifier))
	if err != nil {
		checks = append(checks, report.Fail(report.CheckOIDCToken, "token_endpoint", redactor.String(err.Error()), nil))
		return Result{Checks: checks}
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		checks = append(checks, report.Fail(report.CheckOIDCToken, "token_endpoint", "id_token missing from token response", nil))
		return Result{Checks: checks}
	}
	checks = append(checks, report.Pass(report.CheckOIDCToken, map[string]string{"id_token_fingerprint": redact.Fingerprint(rawIDToken)}))

	claims, verifyCheck := VerifyIDToken(ctx, provider, rawIDToken, cfg.IDP.ClientID, nonce)
	checks = append(checks, verifyCheck)
	if verifyCheck.Status != report.StatusPass {
		return Result{IDToken: rawIDToken, Checks: checks}
	}
	claimCheck := ValidateClaims(claims, cfg.ExpectedClaims)
	checks = append(checks, claimCheck)
	return Result{IDToken: rawIDToken, Claims: claims, Checks: checks}
}

type callbackResult struct {
	Code string
}

func waitForCallback(ctx context.Context, cfg config.Callback, expectedState string) (callbackResult, report.Check) {
	bind := cfg.Bind
	if bind == "" {
		bind = "127.0.0.1"
	}
	addr := fmt.Sprintf("%s:%d", bind, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return callbackResult{}, report.Fail(report.CheckOIDCCallback, "callback_listener", err.Error(), nil)
	}
	defer listener.Close()

	server := &http.Server{ReadHeaderTimeout: 5 * time.Second}
	resultCh := make(chan callbackResult, 1)
	checkCh := make(chan report.Check, 1)
	path := cfg.Path
	if path == "" {
		path = "/callback"
	}
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if errValue := query.Get("error"); errValue != "" {
			checkCh <- report.Fail(report.CheckOIDCCallback, "authorization_error", errValue, nil)
			http.Error(w, "authorization failed", http.StatusBadRequest)
			return
		}
		if subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(expectedState)) != 1 {
			checkCh <- report.Fail(report.CheckOIDCCallback, "state_mismatch", "callback state did not match", nil)
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		code := query.Get("code")
		if code == "" {
			checkCh <- report.Fail(report.CheckOIDCCallback, "missing_code", "authorization code missing from callback", nil)
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		resultCh <- callbackResult{Code: code}
		checkCh <- report.Pass(report.CheckOIDCCallback, map[string]string{"path": path})
		_, _ = w.Write([]byte("SSO test callback received. You can close this window."))
	})
	server.Handler = mux
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Shutdown(context.Background())

	timeout := cfg.TimeoutDuration()
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return callbackResult{}, report.Fail(report.CheckOIDCCallback, "canceled", ctx.Err().Error(), nil)
	case <-timer.C:
		return callbackResult{}, report.Fail(report.CheckOIDCCallback, "timeout", "callback timed out", nil)
	case check := <-checkCh:
		if check.Status != report.StatusPass {
			return callbackResult{}, check
		}
		return <-resultCh, check
	}
}

func VerifyIDToken(ctx context.Context, provider *oidc.Provider, rawIDToken, clientID, nonce string) (map[string]any, report.Check) {
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, report.Fail(report.CheckOIDCIDToken, "id_token_validation", err.Error(), nil)
	}
	if nonce != "" && idToken.Nonce != nonce {
		return nil, report.Fail(report.CheckOIDCIDToken, "nonce_mismatch", "id token nonce did not match", nil)
	}
	claims := map[string]any{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, report.Fail(report.CheckOIDCIDToken, "claims_decode", err.Error(), nil)
	}
	return claims, report.Pass(report.CheckOIDCIDToken, map[string]string{"token_fingerprint": redact.Fingerprint(rawIDToken)})
}

func ValidateClaims(claims map[string]any, rules []config.ClaimRule) report.Check {
	for _, rule := range rules {
		value, ok := claims[rule.Name]
		if rule.Required && !ok {
			return report.Fail(report.CheckOIDCClaims, "missing_claim", fmt.Sprintf("claim %s is required", rule.Name), nil)
		}
		if !ok {
			continue
		}
		if rule.ExpectedValue != "" && fmt.Sprint(value) != rule.ExpectedValue {
			return report.Fail(report.CheckOIDCClaims, "claim_mismatch", fmt.Sprintf("claim %s did not match expected value", rule.Name), nil)
		}
		if len(rule.AllowedValues) > 0 && !contains(rule.AllowedValues, fmt.Sprint(value)) {
			return report.Fail(report.CheckOIDCClaims, "claim_mismatch", fmt.Sprintf("claim %s was not an allowed value", rule.Name), nil)
		}
	}
	return report.Pass(report.CheckOIDCClaims, map[string]string{"validated_claims": fmt.Sprintf("%d", len(rules))})
}

func RandomURLSafe(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if strings.EqualFold(value, needle) {
			return true
		}
	}
	return false
}
