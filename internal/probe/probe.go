package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func Run(ctx context.Context, cfg config.Probe, resolvedSecrets map[string]string, accessToken string, client HTTPClient, redactor redact.Redactor) []report.Check {
	if !cfg.Enabled {
		return []report.Check{report.Skip(report.CheckProbeReq, "probe disabled")}
	}
	req, err := http.NewRequestWithContext(ctx, method(cfg.Method), cfg.URL, nil)
	if err != nil {
		return []report.Check{report.Fail(report.CheckProbeReq, "probe_request", err.Error(), nil)}
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	for key, value := range cfg.Headers {
		if resolved, ok := resolvedSecrets["probe.headers."+key]; ok {
			value = resolved
		}
		req.Header.Set(key, value)
	}

	checks := []report.Check{report.Pass(report.CheckProbeReq, map[string]string{"url": redactor.URL(cfg.URL), "method": req.Method})}
	resp, err := client.Do(req)
	if err != nil {
		checks = append(checks, report.Fail(report.CheckProbeResp, "transport", redactor.String(err.Error()), nil))
		return checks
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if readErr != nil {
		checks = append(checks, report.Fail(report.CheckProbeResp, "read_response", readErr.Error(), nil))
		return checks
	}

	evidence := map[string]string{"status": resp.Status, "response_code": fmt.Sprintf("%d", resp.StatusCode)}
	if cfg.RedactBody {
		evidence["body_summary"] = redactor.String(string(body))
	}
	if expected(resp.StatusCode, cfg.ExpectedStatus) {
		checks = append(checks, report.Pass(report.CheckProbeResp, evidence))
		return checks
	}

	checks = append(checks, report.Fail(report.CheckProbeResp, classify(resp.StatusCode), resp.Status, evidence))
	return checks
}

func method(value string) string {
	if strings.TrimSpace(value) == "" {
		return http.MethodGet
	}
	return value
}

func expected(status int, allowed []int) bool {
	if len(allowed) == 0 {
		return status >= 200 && status <= 299
	}
	for _, candidate := range allowed {
		if candidate == status {
			return true
		}
	}
	return false
}

func classify(status int) string {
	switch {
	case status == http.StatusUnauthorized:
		return "unauthorized"
	case status == http.StatusForbidden:
		return "forbidden"
	case status == http.StatusNotFound:
		return "not_found"
	case status == http.StatusTooManyRequests:
		return "rate_limited"
	case status >= 500:
		return "server_error"
	default:
		return "unexpected_status"
	}
}
