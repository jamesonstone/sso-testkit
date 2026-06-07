package report

import (
	"bytes"
	"encoding/json"
	"sort"
	"time"
)

type Status string

const (
	StatusPass    Status = "pass"
	StatusFail    Status = "fail"
	StatusBlocked Status = "blocked"
	StatusSkip    Status = "skip"
)

const (
	CheckConfigLoad     = "config.load"
	CheckConfigValidate = "config.validate"
	CheckOIDCDiscovery  = "oidc.discovery"
	CheckOIDCAuthURL    = "oidc.auth_url"
	CheckOIDCCallback   = "oidc.callback"
	CheckOIDCToken      = "oidc.token"
	CheckOIDCIDToken    = "oidc.id_token"
	CheckOIDCClaims     = "oidc.claims"
	CheckExchangeReq    = "exchange.request"
	CheckExchangeResp   = "exchange.response"
	CheckProbeReq       = "probe.request"
	CheckProbeResp      = "probe.response"
	CheckRedactionScan  = "redaction.scan"
)

type Check struct {
	Name              string            `json:"name"`
	Status            Status            `json:"status"`
	Evidence          map[string]string `json:"evidence,omitempty"`
	FailureReason     string            `json:"failure_reason,omitempty"`
	Classification    string            `json:"classification,omitempty"`
	Owner             string            `json:"owner,omitempty"`
	RecommendedAction string            `json:"recommended_action,omitempty"`
}

type Report struct {
	Version    string    `json:"version"`
	ScenarioID string    `json:"scenario_id"`
	Mode       string    `json:"mode"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Overall    Status    `json:"overall_status"`
	Checks     []Check   `json:"checks"`
}

func New(scenarioID, mode string, startedAt time.Time) Report {
	return Report{
		Version:    "1",
		ScenarioID: scenarioID,
		Mode:       mode,
		StartedAt:  startedAt.UTC(),
		Overall:    StatusSkip,
	}
}

func (r *Report) Add(check Check) {
	if check.Evidence != nil && len(check.Evidence) == 0 {
		check.Evidence = nil
	}
	r.Checks = append(r.Checks, check)
}

func (r *Report) Finish(finishedAt time.Time) {
	r.FinishedAt = finishedAt.UTC()
	r.Overall = OverallStatus(r.Checks)
}

func OverallStatus(checks []Check) Status {
	if len(checks) == 0 {
		return StatusSkip
	}

	seenPass := false
	seenBlocked := false
	for _, check := range checks {
		switch check.Status {
		case StatusFail:
			return StatusFail
		case StatusBlocked:
			seenBlocked = true
		case StatusPass:
			seenPass = true
		}
	}

	if seenBlocked {
		return StatusBlocked
	}
	if seenPass {
		return StatusPass
	}
	return StatusSkip
}

func (r Report) SortedChecks() []Check {
	checks := append([]Check(nil), r.Checks...)
	sort.SliceStable(checks, func(i, j int) bool {
		return checks[i].Name < checks[j].Name
	})
	return checks
}

func (r Report) JSON() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func Pass(name string, evidence map[string]string) Check {
	return Check{Name: name, Status: StatusPass, Evidence: evidence}
}

func Skip(name, reason string) Check {
	return Check{Name: name, Status: StatusSkip, FailureReason: reason}
}

func Fail(name, classification, reason string, evidence map[string]string) Check {
	return Check{Name: name, Status: StatusFail, Classification: classification, FailureReason: reason, Evidence: evidence}
}

func Blocked(name, classification, reason, owner, action string, evidence map[string]string) Check {
	return Check{
		Name:              name,
		Status:            StatusBlocked,
		Classification:    classification,
		FailureReason:     reason,
		Owner:             owner,
		RecommendedAction: action,
		Evidence:          evidence,
	}
}
