---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
parallelization_mode: rlm
references:
  - id: spec
    name: Init project spec
    type: feature_doc
    target: docs/specs/0001-init-project/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding feature contract, acceptance criteria, and non-goals
    status: active
  - id: brainstorm
    name: Init project brainstorm
    type: feature_doc
    target: docs/specs/0001-init-project/BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, accepted defaults, and initial file-surface findings
    status: active
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: project-wide constraints and formal workflow classification
    status: active
  - id: agent-workflows
    name: Agent workflows
    type: repo_doc
    target: docs/agents/WORKFLOWS.md
    relation: guides
    read_policy: must
    used_for: source-of-truth order and spec-driven workflow rules
    status: active
  - id: agent-guardrails
    name: Agent guardrails
    type: repo_doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion bar, safety rules, and validation expectations
    status: active
  - id: agent-rlm
    name: RLM context routing
    type: repo_doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: just-in-time context loading and prior-work filtering
    status: active
  - id: kit-map-init-project
    name: kit map 0001-init-project
    type: command
    target: "kit map 0001-init-project"
    selector_type: command
    selector: "kit map 0001-init-project"
    relation: informs
    read_policy: evidence
    used_for: confirming current phase and absence of feature relationships
    status: active
  - id: project-progress
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: informs
    read_policy: conditional
    used_for: confirming feature phase and open-item state
    status: active
  - id: yp-readme
    name: yp README
    type: local_reference
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md
    relation: informs
    read_policy: evidence
    used_for: README header-style, tagline, quick start, requirements, and development section pattern
    status: active
  - id: bwell-token-exchange
    name: b.well OAuth Token Exchange
    type: url
    target: "https://developer.bwell.com/docs/oauth-token-exchange"
    relation: informs
    read_policy: evidence
    used_for: first service-provider token exchange strategy
    status: active
  - id: rfc-8693
    name: RFC 8693 OAuth 2.0 Token Exchange
    type: url
    target: "https://www.rfc-editor.org/rfc/rfc8693"
    relation: constrains
    read_policy: must
    used_for: generic token exchange request and response semantics
    status: active
  - id: oidc-core
    name: OpenID Connect Core 1.0
    type: url
    target: "https://openid.net/specs/openid-connect-core-1_0.html"
    relation: constrains
    read_policy: must
    used_for: ID token claim and validation strategy
    status: active
  - id: oidc-discovery
    name: OpenID Connect Discovery 1.0
    type: url
    target: "https://openid.net/specs/openid-connect-discovery-1_0.html"
    relation: constrains
    read_policy: must
    used_for: provider metadata and JWKS discovery strategy
    status: active
  - id: rfc-7636
    name: RFC 7636 PKCE
    type: url
    target: "https://www.rfc-editor.org/rfc/rfc7636"
    relation: constrains
    read_policy: must
    used_for: PKCE verifier and challenge strategy
    status: active
  - id: rfc-7517
    name: RFC 7517 JSON Web Key
    type: url
    target: "https://www.rfc-editor.org/rfc/rfc7517"
    relation: constrains
    read_policy: must
    used_for: JWKS-backed token validation strategy
    status: active
  - id: auth0-oidc
    name: Auth0 OpenID Connect Protocol
    type: url
    target: "https://auth0.com/docs/authenticate/protocols/openid-connect-protocol"
    relation: informs
    read_policy: evidence
    used_for: first identity-provider scenario assumptions
    status: active
---
# PLAN

## SUMMARY

Create a small Go module with a single `sso-testkit` binary that loads a strict YAML scenario, validates configuration before network calls, runs either a stubbed readiness path or a live OIDC callback flow, performs token exchange and optional bearer probing through protocol-oriented packages, and writes a deterministic redacted report. The design keeps Auth0 and b.well in configuration and thin adapters, while core packages stay reusable for other IdP/SP scenarios.

## APPROACH

1. Start with a standard Go module and a narrow command surface rather than a web framework. Use `cmd/sso-testkit` for the binary and keep reusable behavior under `internal/` packages.
2. Use standard library packages for CLI flags, HTTP server/client behavior, context cancellation, random bytes, hashing, JSON output, and structured logging. Add mature dependencies only for OAuth2, OIDC/JWKS validation, and strict YAML decoding so security-sensitive protocol behavior is not hand-rolled.
3. Make configuration validation the first runtime phase. The command should load YAML, reject unknown fields, resolve environment references, classify missing values, and stop before network calls when the scenario is invalid.
4. Treat scenarios as the provider plug-in boundary. Auth0 and b.well should appear in `configs/scenarios/bwell-config.yaml` and parsed config values, not as hard-coded branches in the main orchestration path.
5. Model protocol steps as check runners that append results to a report. Each runner receives already-validated config, a redactor, and an HTTP client, then returns a typed result with status, evidence, failure reason, and owner/action metadata.
6. Split live and stub behavior at the runner boundary. Stub mode should exercise config parsing, callback/state/nonce helpers, token exchange request construction, probe classification, report rendering, and redaction without external credentials.
7. Keep b.well-specific exchange behavior as a thin exchange profile that maps scenario values into the generic RFC 8693 form request plus configured headers, including `clientkey`.
8. Prefer deterministic reports over rich terminal UX. Terminal output can summarize report status, but JSON report generation should be the primary verification artifact.
9. Write README content against real commands chosen here: build, test, validate a scenario, run stub readiness, and run live readiness.

Tradeoffs:

1. Use `flag` and explicit subcommand dispatch instead of a CLI framework. This keeps dependency surface small for a first release; a framework can be added only if command complexity grows.
2. Use in-memory callback state instead of persistent storage. The first release is local and single-run; a database would add operational weight without satisfying additional requirements.
3. Use `github.com/coreos/go-oidc/v3/oidc`, `golang.org/x/oauth2`, and `gopkg.in/yaml.v3` as the likely non-standard dependencies. These are focused enough to avoid broad frameworks while reducing risk in protocol and parser code.
4. Do not auto-open the browser by default. Print the authorization URL and optionally support an `--open-browser` flag later if implementation remains simple; this avoids cross-platform behavior becoming a first-release blocker.
5. Commit only safe scenario templates. Real local scenario files with secret-bearing values should be excluded by pattern or documented as untracked local copies.

## COMPONENTS

1. `cmd/sso-testkit`
   - Owns process entry, subcommand parsing, exit codes, stdout/stderr boundaries, and top-level context cancellation.
   - Delegates all protocol, config, reporting, and redaction behavior to `internal/` packages.
2. `internal/app`
   - Orchestrates command flows such as `validate-config`, `run --mode stub`, and `run --mode live`.
   - Converts typed check outcomes into process exit status.
3. `internal/config`
   - Loads YAML from a selected path.
   - Performs strict field validation, required-field validation, mode validation, URL validation, duration validation, and secret-reference validation.
   - Resolves environment references only after syntax validation and before runners execute.
4. `internal/secretref`
   - Represents secret sources such as environment variables.
   - Keeps secret resolution separate from general config parsing so redaction and validation can reason about sensitive fields consistently.
5. `internal/redact`
   - Centralizes token, secret, header, URL query, and error-body redaction.
   - Produces stable non-reversible fingerprints for diagnostics.
6. `internal/report`
   - Defines check status values, check names, evidence summaries, failure classifications, owner/action hints, and JSON rendering.
   - Provides deterministic ordering for checks and fields.
7. `internal/oidcflow`
   - Handles discovery, PKCE/state/nonce generation, local callback handling, token exchange with the IdP token endpoint, and ID token validation.
   - Wraps OIDC library behavior behind local types that return reportable check outcomes.
8. `internal/exchange`
   - Builds and executes generic RFC 8693 token exchange requests.
   - Supports stub responses and provider-specific exchange profiles without leaking provider branching into the app runner.
9. `internal/probe`
   - Executes configured bearer/API probes with the exchanged access token.
   - Classifies HTTP status, timeout, TLS, malformed response, and redacted body summaries.
10. `internal/testsupport`
   - Provides local fake OIDC/exchange/probe servers, sample tokens, and report helpers for tests only.
   - Prevents tests from needing live Auth0 or b.well credentials.
11. `configs/scenarios`
   - Holds safe committed scenario templates, starting with `bwell-config.yaml`.
   - Uses environment references for secrets and placeholders for tenant-specific values.
12. `README.md`
   - Documents actual commands, the scenario model, redaction guarantees, requirements, and development checks using the `yp` README style.

## DATA

1. Scenario file
   - Format: YAML.
   - First committed path: `configs/scenarios/bwell-config.yaml`.
   - Top-level shape: `scenario`, `mode_defaults`, `callback`, `idp`, `expected_claims`, `exchange`, `probe`, `report`, and `redaction`.
2. Scenario identity
   - Fields: stable scenario ID, display name, description, and provider labels.
   - Provider labels are descriptive only; behavior should come from protocol/profile fields.
3. Mode
   - Values: `stub` and `live`.
   - CLI mode overrides scenario default; invalid combinations fail config validation.
4. Callback settings
   - Fields: bind address, port, path, redirect URI, success page text, timeout.
   - Local callback state remains in memory for one run.
5. IdP settings
   - Fields: issuer URL, client ID, optional client secret reference, scopes, redirect URI override, audience/resource values when needed, expected signing algorithms where supported.
   - Auth0 tenant specifics live here in the b.well scenario template.
6. Expected claims
   - Shape: list or map of claim rules with claim name, required flag, expected value or allowed values, and redaction classification.
   - Claim values from tokens are never persisted raw in reports.
7. Exchange settings
   - Fields: profile (`rfc8693` or `bwell`), endpoint URL, method, headers, subject token source, subject token type, requested token type, optional resource/audience/scope, timeout, and stub response.
   - Secret header values use secret references.
8. Probe settings
   - Fields: enabled flag, URL, method, headers, expected status set, timeout, and body redaction policy.
   - Probe body capture should be opt-in and redacted.
9. Secret reference
   - Preferred syntax: `env:VARIABLE_NAME`.
   - Literal values are allowed only for explicitly non-sensitive fields.
10. Report
   - Fields: scenario ID, mode, started/finished timestamps, overall status, ordered checks, redacted evidence, failure classification, owner/action, and version metadata.
   - Status values: `pass`, `fail`, `blocked`, `skip`.
11. Check names
   - Stable names should include `config.load`, `config.validate`, `oidc.discovery`, `oidc.auth_url`, `oidc.callback`, `oidc.token`, `oidc.id_token`, `oidc.claims`, `exchange.request`, `exchange.response`, `probe.request`, `probe.response`, and `redaction.scan`.

## INTERFACES

1. Command: `sso-testkit validate-config --config <path> [--mode stub|live]`
   - Input: scenario YAML path and optional mode override.
   - Output: validation summary and non-zero exit on invalid config.
   - Side effects: no network calls, no report file unless explicitly requested later.
2. Command: `sso-testkit run --config <path> --mode stub [--report <path>]`
   - Input: scenario YAML path, stub mode, optional report path.
   - Output: deterministic redacted readiness report and terminal summary.
   - Side effects: local-only stub servers/helpers as needed, no external Auth0 or b.well calls.
3. Command: `sso-testkit run --config <path> --mode live [--report <path>]`
   - Input: scenario YAML path, live mode, environment-provided secrets, optional report path.
   - Output: redacted readiness report and terminal summary.
   - Side effects: local callback listener, external IdP requests, external exchange/probe requests according to scenario.
4. Default report output
   - If `--report` is omitted, write a terminal summary and optionally a deterministic JSON report under a documented local output directory only if the implementation chooses that behavior consistently.
   - If `--report -` is supported, write JSON to stdout and send human logs to stderr.
5. Exit codes
   - `0`: all required checks passed or only non-required checks skipped.
   - `1`: readiness failed or was blocked.
   - `2`: command/config usage error before checks ran.
6. Files created during implementation
   - `go.mod`, `go.sum`, `cmd/sso-testkit/main.go`, `internal/**`, `configs/scenarios/bwell-config.yaml`, `README.md`.
   - Optional local output directory may be documented and ignored if report files are written by default.
7. Files not touched during implementation
   - Do not edit generated or Kit docs except when implementation materially changes the accepted plan/spec.
8. Observable evidence
   - Each acceptance item must map to command output, unit/integration test assertions, `go test ./...`, `go vet ./...`, `git diff --check`, or a manually gated live b.well run.

## DEPENDENCIES

References are tracked in front matter.

Implementation-shaping dependencies:

1. Go standard library: `context`, `crypto/rand`, `crypto/sha256`, `encoding/json`, `errors`, `flag`, `fmt`, `log/slog`, `net/http`, `net/url`, `os`, `os/signal`, `strings`, and `time`.
2. OAuth/OIDC libraries: `golang.org/x/oauth2` and `github.com/coreos/go-oidc/v3/oidc`, unless implementation discovery finds a safer maintained equivalent.
3. YAML library: `gopkg.in/yaml.v3` with strict known-field decoding.
4. External standards: OIDC Core, OIDC Discovery, RFC 7636, RFC 7517, and RFC 8693.
5. First provider references: Auth0 OIDC docs and b.well OAuth Token Exchange docs.
6. Documentation style reference: `/Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md`.

## RISKS

1. Risk: accidental token or secret leakage through logs, reports, errors, fixtures, or failed HTTP bodies.
   - Mitigation: route all diagnostics through `internal/redact`, add redaction tests using realistic token/secret samples, and keep report evidence summaries allowlisted rather than dumping raw structs.
2. Risk: hand-rolled OIDC/JWKS validation introduces security bugs.
   - Mitigation: use a mature OIDC library for discovery and ID token verification; keep local code focused on orchestration and report classification.
3. Risk: YAML scenario flexibility becomes an untyped mini-language.
   - Mitigation: define a strict first-release schema, reject unknown fields, and support only the protocol/profile fields needed by the SPEC.
4. Risk: b.well-specific behavior contaminates generic token exchange.
   - Mitigation: isolate b.well differences in an exchange profile that only maps config into headers and RFC 8693 form parameters.
5. Risk: live external dependencies make tests flaky or impossible for contributors.
   - Mitigation: default tests use fake servers and stub mode; live checks require explicit environment variables and are documented as manual or opt-in.
6. Risk: local callback listener hangs or leaves resources open.
   - Mitigation: use context deadlines, one-shot callback handling, graceful server shutdown, and tests for timeout/cancel paths.
7. Risk: readiness reports vary across runs and become hard to test.
   - Mitigation: sort checks and keys, inject clock/test time where needed, and normalize dynamic values behind redacted fingerprints.
8. Risk: README commands drift from real CLI behavior.
   - Mitigation: finalize README after command names are implemented and include a smoke check that exercises documented commands.

## TESTING

1. Config validation tests
   - Evidence: `go test ./internal/config ./internal/secretref` passes.
   - Coverage: valid b.well template, unknown fields, missing required fields, invalid URLs, invalid modes, unset environment references, and literal secret rejection.
2. Redaction tests
   - Evidence: tests prove sample ID tokens, access tokens, refresh tokens, client secrets, client keys, authorization headers, and sensitive query values do not appear in rendered logs/reports/errors.
   - Coverage: both success and failure paths.
3. OIDC flow tests
   - Evidence: `go test ./internal/oidcflow` passes against local fake OIDC metadata, JWKS, token endpoint, and callback server.
   - Coverage: state mismatch, nonce mismatch, missing code, user cancel, issuer mismatch, audience mismatch, expired token, invalid signature, missing expected claim, and happy-path stub login.
4. Exchange tests
   - Evidence: `go test ./internal/exchange` proves RFC 8693 request construction and b.well profile header handling without raw token output.
   - Coverage: stub response, HTTP error, malformed response, timeout, and b.well sandbox-trust failure classification.
5. Probe tests
   - Evidence: `go test ./internal/probe` passes against local fake HTTP servers.
   - Coverage: expected status, 401, 403, 404, 429, 5xx, timeout, TLS or transport errors where practical, and redacted body summaries.
6. Report tests
   - Evidence: `go test ./internal/report` proves deterministic JSON output and stable check ordering.
   - Coverage: pass, fail, blocked, skip, owner/action hints, and timestamp normalization in tests.
7. CLI/app integration tests
   - Evidence: `go test ./cmd/sso-testkit ./internal/app` or equivalent subprocess tests pass.
   - Coverage: `validate-config`, `run --mode stub`, bad arguments, exit codes, report-to-file, and report-to-stdout if supported.
8. Repository-level validation
   - Evidence: `go test ./...`, `go vet ./...`, and `git diff --check` pass after implementation.
9. README smoke validation
   - Evidence: documented build/test/validate/stub commands run successfully in a clean checkout with no live credentials.
10. Live b.well validation
   - Evidence: explicit manual or opt-in command documents the environment variables used and captures a redacted report.
   - Scope: live success is not required for default CI; incomplete b.well trust must produce a clear `blocked` or `fail` report without claiming downstream account/consent/search success.
