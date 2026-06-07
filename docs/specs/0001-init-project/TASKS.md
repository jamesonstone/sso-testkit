---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Initialize Go module, dependency baseline, and package directories [PLAN-APPROACH](#approach) | done | agent | |
| T002 | Define report status model and deterministic JSON rendering [PLAN-DATA](#data) | done | agent | T001 |
| T003 | Implement centralized redaction and secret reference primitives [PLAN-COMPONENTS](#components) | done | agent | T001, T002 |
| T004 | Implement strict YAML scenario schema and validation [PLAN-DATA](#data) | done | agent | T003 |
| T005 | Add safe b.well scenario template [PLAN-DATA](#data) | done | agent | T004 |
| T006 | Implement CLI/app command orchestration and exit codes [PLAN-INTERFACES](#interfaces) | done | agent | T002, T004 |
| T007 | Implement OIDC PKCE, callback, discovery, and token validation flow [PLAN-COMPONENTS](#components) | done | agent | T006 |
| T008 | Implement RFC 8693 exchange runner and b.well profile [PLAN-COMPONENTS](#components) | done | agent | T006 |
| T009 | Implement bearer/API probe runner [PLAN-COMPONENTS](#components) | done | agent | T006, T008 |
| T010 | Wire stub and live run modes into app workflow [PLAN-APPROACH](#approach) | done | agent | T007, T008, T009 |
| T011 | Add fake-server and test-support utilities [PLAN-TESTING](#testing) | done | agent | T007, T008, T009 |
| T012 | Complete package and CLI tests for required acceptance paths [PLAN-TESTING](#testing) | done | agent | T010, T011 |
| T013 | Add populated README with verified commands [PLAN-INTERFACES](#interfaces) | done | agent | T010, T012 |
| T014 | Run repository validation and close readiness gaps [PLAN-TESTING](#testing) | done | agent | T012, T013 |

## TASK LIST

- [x] T001: Initialize Go module, dependency baseline, and package directories [PLAN-APPROACH](#approach)
- [x] T002: Define report status model and deterministic JSON rendering [PLAN-DATA](#data)
- [x] T003: Implement centralized redaction and secret reference primitives [PLAN-COMPONENTS](#components)
- [x] T004: Implement strict YAML scenario schema and validation [PLAN-DATA](#data)
- [x] T005: Add safe b.well scenario template [PLAN-DATA](#data)
- [x] T006: Implement CLI/app command orchestration and exit codes [PLAN-INTERFACES](#interfaces)
- [x] T007: Implement OIDC PKCE, callback, discovery, and token validation flow [PLAN-COMPONENTS](#components)
- [x] T008: Implement RFC 8693 exchange runner and b.well profile [PLAN-COMPONENTS](#components)
- [x] T009: Implement bearer/API probe runner [PLAN-COMPONENTS](#components)
- [x] T010: Wire stub and live run modes into app workflow [PLAN-APPROACH](#approach)
- [x] T011: Add fake-server and test-support utilities [PLAN-TESTING](#testing)
- [x] T012: Complete package and CLI tests for required acceptance paths [PLAN-TESTING](#testing)
- [x] T013: Add populated README with verified commands [PLAN-INTERFACES](#interfaces)
- [x] T014: Run repository validation and close readiness gaps [PLAN-TESTING](#testing)

## TASK DETAILS

### T001

- GOAL: Establish the initial Go project structure and dependency baseline.
- SCOPE:
  - Create `go.mod` and `go.sum`.
  - Add `cmd/sso-testkit` and planned `internal/` package directories.
  - Add only focused OAuth2, OIDC, and YAML dependencies approved by PLAN.
  - Keep generated or placeholder code minimal and buildable.
- ACCEPTANCE:
  - `go list ./...` succeeds.
  - `go mod tidy` produces a clean module state.
  - No application behavior is hard-coded to Auth0 or b.well.
- EVIDENCE:
  - `go list ./...`
  - `go mod tidy`
  - `git diff -- go.mod go.sum cmd internal`
- NOTES: no additional information required

### T002

- GOAL: Create the report data model and deterministic JSON rendering used by all readiness checks.
- SCOPE:
  - Define status values `pass`, `fail`, `blocked`, and `skip`.
  - Define stable check names from PLAN.
  - Add report, check result, evidence, failure classification, owner/action, and version metadata shapes.
  - Make JSON output deterministic with stable ordering and testable timestamps.
- ACCEPTANCE:
  - Report rendering produces stable JSON for identical input.
  - Report model can represent config, OIDC, exchange, probe, and redaction checks.
  - Package tests cover all statuses.
- EVIDENCE:
  - `go test ./internal/report`
  - Golden or explicit JSON assertion in report tests
- NOTES: no additional information required

### T003

- GOAL: Provide a single safe path for secret references and redacted diagnostics.
- SCOPE:
  - Implement `env:VARIABLE_NAME` secret references.
  - Reject malformed secret references.
  - Add token, secret, header, query, and error-body redaction helpers.
  - Add stable non-reversible fingerprints for diagnostics.
- ACCEPTANCE:
  - Raw sample tokens, client secrets, client keys, authorization headers, and sensitive query values are redacted in tests.
  - Unset environment variables are reported by variable name only.
  - Redaction helpers are reusable by config, report, exchange, probe, and app code.
- EVIDENCE:
  - `go test ./internal/secretref ./internal/redact`
  - Test fixtures proving raw sensitive samples do not appear in rendered output
- NOTES: no additional information required

### T004

- GOAL: Implement strict YAML scenario parsing and validation before any network calls.
- SCOPE:
  - Define scenario schema for `scenario`, `mode_defaults`, `callback`, `idp`, `expected_claims`, `exchange`, `probe`, `report`, and `redaction`.
  - Reject unknown fields, missing required fields, invalid URLs, invalid modes, invalid durations, and literal secrets in sensitive fields.
  - Resolve secret references only after structural validation.
  - Return actionable validation errors.
- ACCEPTANCE:
  - Valid scenario fixtures parse successfully.
  - Unknown fields, missing required values, malformed URLs, invalid modes, unset env refs, and literal sensitive values fail before network calls.
  - Config tests do not require live credentials.
- EVIDENCE:
  - `go test ./internal/config ./internal/secretref`
  - Fixture coverage for valid and invalid scenario files
- NOTES: no additional information required

### T005

- GOAL: Add the first safe Auth0-to-b.well scenario template.
- SCOPE:
  - Create `configs/scenarios/bwell-config.yaml`.
  - Model Auth0 as IdP and b.well as SP through scenario fields.
  - Use environment references for all secrets and keys.
  - Avoid committed literal credentials or raw tokens.
  - Add ignore/documentation handling for local secret-bearing scenario copies if needed.
- ACCEPTANCE:
  - Template passes `validate-config` or config package validation when required non-secret placeholders/env references are supplied.
  - Secret scan or redaction-focused tests confirm no literal credential values are committed.
  - Template keeps provider details in YAML rather than application code.
- EVIDENCE:
  - `go test ./internal/config`
  - `rg -n "client_secret|access_token|id_token|refresh_token|clientkey" configs/scenarios`
  - Manual review of `configs/scenarios/bwell-config.yaml`
- NOTES: no additional information required

### T006

- GOAL: Add the command surface and top-level app orchestration.
- SCOPE:
  - Implement `sso-testkit validate-config --config <path> [--mode stub|live]`.
  - Implement `sso-testkit run --config <path> --mode stub|live [--report <path>]`.
  - Implement exit codes `0`, `1`, and `2` as defined in PLAN.
  - Keep human logs and JSON report output separated when report-to-stdout is supported.
  - Delegate protocol work to internal packages.
- ACCEPTANCE:
  - Bad arguments return usage/config exit behavior.
  - `validate-config` performs no network calls.
  - CLI tests can exercise command parsing and exit codes without live credentials.
- EVIDENCE:
  - `go test ./cmd/sso-testkit ./internal/app`
  - CLI test assertions for usage, invalid config, and valid config
- NOTES: no additional information required

### T007

- GOAL: Implement the OIDC readiness runner for PKCE, callback handling, discovery, and ID token validation.
- SCOPE:
  - Generate PKCE verifier/challenge, state, and nonce.
  - Discover provider metadata and JWKS from issuer.
  - Run local one-shot callback listener for live mode.
  - Validate ID token signature, issuer, audience, expiration, issued-at tolerance, nonce, and configured claims.
  - Return typed report check results for each OIDC step.
- ACCEPTANCE:
  - Happy-path fake OIDC flow passes in tests.
  - State mismatch, nonce mismatch, missing code, user cancel, issuer mismatch, audience mismatch, expired token, invalid signature, and missing expected claim are rejected.
  - No raw ID token appears in report or logs.
- EVIDENCE:
  - `go test ./internal/oidcflow`
  - Redaction assertions covering OIDC failure paths
- NOTES: no additional information required

### T008

- GOAL: Implement generic RFC 8693 token exchange and the thin b.well exchange profile.
- SCOPE:
  - Build RFC 8693 form requests from scenario values.
  - Support subject token source, subject token type, requested token type, resource/audience/scope, timeout, and headers.
  - Add b.well profile mapping for configured endpoint and `clientkey` header behavior.
  - Support stub exchange responses.
  - Classify HTTP, timeout, malformed response, and sandbox-trust failures.
- ACCEPTANCE:
  - Tests prove request construction includes required RFC 8693 fields.
  - b.well profile tests prove `clientkey` handling without hard-coded credentials.
  - Raw subject token and exchanged token values are redacted in all outputs.
- EVIDENCE:
  - `go test ./internal/exchange`
  - Request-construction assertions from fake exchange server
- NOTES: no additional information required

### T009

- GOAL: Implement downstream bearer/API probing with redacted response classification.
- SCOPE:
  - Build probe requests from scenario config.
  - Apply exchanged access token as bearer token when configured.
  - Classify expected status, 401, 403, 404, 429, 5xx, timeout, TLS/transport errors, and malformed responses.
  - Redact headers, query strings, and response summaries.
- ACCEPTANCE:
  - Fake-server tests cover success and failure status classes.
  - Probe failures are distinguishable from authentication and exchange failures in reports.
  - No bearer token appears in captured output.
- EVIDENCE:
  - `go test ./internal/probe`
  - Redacted report assertions for failure cases
- NOTES: no additional information required

### T010

- GOAL: Wire config, OIDC, exchange, probe, report, and redaction into stub and live app flows.
- SCOPE:
  - Implement ordered check execution.
  - Ensure stub mode requires no live Auth0 or b.well credentials.
  - Ensure live mode starts callback listener and uses external endpoints only after config validation.
  - Write report file or stdout according to CLI flags.
  - Convert overall report status to defined exit code.
- ACCEPTANCE:
  - `run --mode stub` produces deterministic report output.
  - `run --mode live` fails early on invalid/missing live config before network calls.
  - Required checks run in stable order.
- EVIDENCE:
  - `go test ./internal/app ./cmd/sso-testkit`
  - Stub-mode report fixture or JSON assertion
- NOTES: no additional information required

### T011

- GOAL: Add reusable local test support for protocol and CLI integration coverage.
- SCOPE:
  - Add fake OIDC discovery/JWKS/token endpoints.
  - Add fake exchange and probe servers.
  - Add sample signed tokens and helpers that remain safe under redaction tests.
  - Add clock or timestamp helpers for deterministic reports.
- ACCEPTANCE:
  - Test support enables package tests without live credentials.
  - Sample sensitive values are covered by redaction assertions.
  - Helpers are isolated under test-only package/files.
- EVIDENCE:
  - `go test ./internal/testsupport ./...`
  - Review that test support is not imported by production packages except in `_test.go`
- NOTES: no additional information required

### T012

- GOAL: Complete acceptance-path tests across packages and CLI behavior.
- SCOPE:
  - Add tests for config, redaction, OIDC, exchange, probe, report, app, and CLI packages.
  - Cover SPEC acceptance criteria and PLAN testing categories.
  - Include negative tests for edge cases listed in SPEC.
  - Keep live b.well checks out of default test suite.
- ACCEPTANCE:
  - `go test ./...` passes without live credentials.
  - Required negative cases fail with typed/reportable outcomes.
  - Redaction tests prove raw sensitive samples are absent from normal outputs.
- EVIDENCE:
  - `go test ./...`
  - Test names or coverage notes mapped to SPEC acceptance items
- NOTES: no additional information required

### T013

- GOAL: Add the project README and verify its commands against implemented behavior.
- SCOPE:
  - Create `README.md` with header-style title block and concise tagline patterned after the `yp` README.
  - Document install/build, quick start, scenario YAML, b.well setup, redaction guarantees, requirements, and development commands.
  - Use only commands that work in the implemented project.
  - Avoid literal secrets or raw token examples.
- ACCEPTANCE:
  - README commands for build, test, validate-config, and stub run execute successfully.
  - README explains live b.well mode as opt-in/manual and credential-dependent.
  - README contains no raw token or secret examples.
- EVIDENCE:
  - README smoke commands from a clean checkout state
  - `rg -n "eyJ|client_secret|refresh_token|access_token" README.md`
- NOTES: no additional information required

### T014

- GOAL: Run final repository validation and resolve any readiness gaps before implementation sign-off.
- SCOPE:
  - Run repository-level Go tests and static checks.
  - Run whitespace diff validation.
  - Run config validation and stub readiness command from README.
  - Run live b.well command only when explicitly configured; otherwise document as not run.
  - Update feature docs only if implementation materially diverged from SPEC or PLAN.
- ACCEPTANCE:
  - `go test ./...` passes.
  - `go vet ./...` passes or documented equivalent passes.
  - `git diff --check` passes.
  - README smoke commands pass.
  - Live b.well validation is either redacted and recorded, or explicitly skipped because credentials/trust are unavailable.
- EVIDENCE:
  - `go test ./...`
  - `go vet ./...`
  - `git diff --check`
  - README smoke command outputs
  - Optional redacted live b.well report
- NOTES: Live b.well validation was not run because credentials and sandbox trust are external; default validation covers stub mode and live-mode preflight behavior.

## DEPENDENCIES

1. T001 must run first because all implementation tasks depend on the Go module and package layout.
2. T002 and T003 establish report/redaction primitives required by every runtime package.
3. T004 and T005 must precede CLI runtime work so app behavior is built around strict scenario config rather than hard-coded provider values.
4. T007, T008, and T009 can be developed after T006 and may proceed in parallel only if shared report/redaction/config contracts remain unchanged.
5. T010 integrates protocol runners and must follow T007 through T009.
6. T011 and T012 complete the fake-server and acceptance coverage after core runners exist.
7. T013 should follow implemented commands so README instructions match real behavior.
8. T014 is the final implementation gate.
9. External blocker: live b.well validation requires user-provided credentials and completed b.well sandbox trust; default implementation and tests must not block on it.

## NOTES

No missing decisions block implementation. If implementation discovery finds that a planned dependency is unsuitable or unmaintained, choose a safer maintained equivalent only when it preserves the SPEC and PLAN contracts.

<!-- REFLECTION_COMPLETE -->
