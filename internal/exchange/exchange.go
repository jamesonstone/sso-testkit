package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
)

const grantTypeTokenExchange = "urn:ietf:params:oauth:grant-type:token-exchange"

type Token struct {
	AccessToken     string
	IssuedTokenType string
	TokenType       string
	ExpiresIn       int
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func Run(ctx context.Context, cfg config.Exchange, resolvedSecrets map[string]string, subjectToken string, client HTTPClient, redactor redact.Redactor, mode string) (Token, []report.Check) {
	if mode == config.ModeStub {
		token := Token{
			AccessToken:     cfg.StubResponse.AccessToken,
			IssuedTokenType: cfg.StubResponse.IssuedTokenType,
			TokenType:       cfg.StubResponse.TokenType,
			ExpiresIn:       cfg.StubResponse.ExpiresIn,
		}
		return token, []report.Check{
			report.Pass(report.CheckExchangeReq, map[string]string{"mode": "stub"}),
			report.Pass(report.CheckExchangeResp, map[string]string{"token_fingerprint": redact.Fingerprint(token.AccessToken)}),
		}
	}

	req, err := BuildRequest(ctx, cfg, resolvedSecrets, subjectToken)
	if err != nil {
		return Token{}, []report.Check{report.Fail(report.CheckExchangeReq, "exchange_request", err.Error(), nil)}
	}
	checks := []report.Check{
		report.Pass(report.CheckExchangeReq, map[string]string{
			"endpoint":           redactor.URL(cfg.Endpoint),
			"subject_token_type": cfg.SubjectTokenType,
			"requested_type":     cfg.RequestedTokenType,
		}),
	}

	resp, err := client.Do(req)
	if err != nil {
		checks = append(checks, report.Fail(report.CheckExchangeResp, "transport", redactor.String(err.Error()), nil))
		return Token{}, checks
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if readErr != nil {
		checks = append(checks, report.Fail(report.CheckExchangeResp, "read_response", readErr.Error(), nil))
		return Token{}, checks
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		classification := "http_error"
		statusText := resp.Status
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			classification = "external_trust"
			checks = append(checks, report.Blocked(report.CheckExchangeResp, classification, statusText, "external_provider", "confirm trust and client configuration", map[string]string{
				"status":        statusText,
				"body_summary":  redactor.String(string(body)),
				"response_code": fmt.Sprintf("%d", resp.StatusCode),
			}))
			return Token{}, checks
		}
		checks = append(checks, report.Fail(report.CheckExchangeResp, classification, statusText, map[string]string{
			"status":        statusText,
			"body_summary":  redactor.String(string(body)),
			"response_code": fmt.Sprintf("%d", resp.StatusCode),
		}))
		return Token{}, checks
	}

	var parsed struct {
		AccessToken     string `json:"access_token"`
		IssuedTokenType string `json:"issued_token_type"`
		TokenType       string `json:"token_type"`
		ExpiresIn       int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		checks = append(checks, report.Fail(report.CheckExchangeResp, "malformed_response", err.Error(), nil))
		return Token{}, checks
	}
	if parsed.AccessToken == "" {
		checks = append(checks, report.Fail(report.CheckExchangeResp, "malformed_response", "access_token missing from exchange response", nil))
		return Token{}, checks
	}

	token := Token{
		AccessToken:     parsed.AccessToken,
		IssuedTokenType: parsed.IssuedTokenType,
		TokenType:       parsed.TokenType,
		ExpiresIn:       parsed.ExpiresIn,
	}
	checks = append(checks, report.Pass(report.CheckExchangeResp, map[string]string{
		"token_type":        token.TokenType,
		"issued_token_type": token.IssuedTokenType,
		"token_fingerprint": redact.Fingerprint(token.AccessToken),
	}))
	return token, checks
}

func BuildRequest(ctx context.Context, cfg config.Exchange, resolvedSecrets map[string]string, subjectToken string) (*http.Request, error) {
	if subjectToken == "" {
		return nil, fmt.Errorf("subject token is required")
	}
	form := url.Values{}
	form.Set("grant_type", grantTypeTokenExchange)
	form.Set("subject_token", subjectToken)
	form.Set("subject_token_type", cfg.SubjectTokenType)
	form.Set("requested_token_type", cfg.RequestedTokenType)
	if cfg.Resource != "" {
		form.Set("resource", cfg.Resource)
	}
	if cfg.Audience != "" {
		form.Set("audience", cfg.Audience)
	}
	if cfg.Scope != "" {
		form.Set("scope", cfg.Scope)
	}

	method := cfg.Method
	if method == "" {
		method = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, method, cfg.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range cfg.Headers {
		field := "exchange.headers." + key
		if resolved, ok := resolvedSecrets[field]; ok {
			req.Header.Set(key, resolved)
			continue
		}
		req.Header.Set(key, value)
	}
	return req, nil
}
