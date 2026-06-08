---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
parallelization_mode: rlm
skills:
  - id: rlm
    name: RLM context routing
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    triggers:
      - analyze codebase
      - scan all files
      - large repository analysis
      - scan repository
      - recursive language model
    used_for: progressive-disclosure discovery and deterministic source-attributed synthesis for broad or noisy repo context
    status: active
relationships: []
references:
  - id: brainstorm
    name: Init project brainstorm
    type: feature_doc
    target: docs/specs/0001-init-project/BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, accepted defaults, and codebase findings
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
    used_for: spec-driven workflow and source-of-truth order
    status: active
  - id: agent-guardrails
    name: Agent guardrails
    type: repo_doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion bar, safety rules, and documentation quality gate
    status: active
  - id: agent-rlm
    name: RLM context routing
    type: repo_doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: just-in-time context loading and prior-work filtering
    status: active
  - id: agent-tooling
    name: Agent tooling
    type: repo_doc
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: conditional
    used_for: skills discovery and secondary input boundaries
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
    used_for: confirming this is the only tracked feature and phase state
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
    used_for: first service-provider preset and token exchange request shape
    status: active
  - id: rfc-8693
    name: RFC 8693 OAuth 2.0 Token Exchange
    type: url
    target: "https://www.rfc-editor.org/rfc/rfc8693"
    relation: constrains
    read_policy: must
    used_for: generic token exchange semantics and token-type identifiers
    status: active
  - id: oidc-core
    name: OpenID Connect Core 1.0
    type: url
    target: "https://openid.net/specs/openid-connect-core-1_0.html"
    relation: constrains
    read_policy: must
    used_for: ID token, issuer, audience, nonce, and claim validation requirements
    status: active
  - id: oidc-discovery
    name: OpenID Connect Discovery 1.0
    type: url
    target: "https://openid.net/specs/openid-connect-discovery-1_0.html"
    relation: constrains
    read_policy: must
    used_for: provider metadata and JWKS URI requirements
    status: active
  - id: rfc-7636
    name: RFC 7636 PKCE
    type: url
    target: "https://www.rfc-editor.org/rfc/rfc7636"
    relation: constrains
    read_policy: must
    used_for: Authorization Code with PKCE flow requirements
    status: active
  - id: rfc-7517
    name: RFC 7517 JSON Web Key
    type: url
    target: "https://www.rfc-editor.org/rfc/rfc7517"
    relation: constrains
    read_policy: must
    used_for: JWKS validation model
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
# SPEC

## SUMMARY

Build `sso-testkit` as a Go-based SSO readiness harness that validates an owned application's OIDC login, token claims, token exchange, and downstream bearer/API reachability before production integration. The first release is protocol-first, redacts secrets and tokens by default, and uses YAML scenario files so Auth0-to-b.well is the first plug-in configuration rather than hard-coded behavior.

## PROBLEM

Integration teams need a small, reusable way to prove that an owned application can complete SSO with an identity provider and then use the resulting identity material with a downstream service provider before production integration work starts. Today, readiness is easy to confuse with app-specific implementation work because identity-provider setup, callback behavior, token validation, token exchange, and downstream API probing are often tested manually or with one-off scripts.

The first concrete scenario is an internal Auth0 identity-provider connection tested against b.well as the service provider. The project must remain useful beyond b.well by keeping provider details in scenario configuration and thin protocol adapters rather than product-specific code paths.

## GOALS

1. Provide a Go application for local SSO readiness testing.
2. Validate OIDC Authorization Code with PKCE against a configured external identity provider.
3. Validate OIDC discovery metadata, JWKS-backed ID token signatures, issuer, audience, nonce, expiration, and configured expected claims.
4. Inspect token/assertion structure only through redacted, safe diagnostics.
5. Support generic RFC 8693 token exchange and generic bearer/API probe readiness checks.
6. Provide a first b.well scenario that exchanges an Auth0 ID token for a b.well access token and reports sandbox trust/configuration failures clearly.
7. Use YAML configuration files per test scenario so external IdP/SP configuration can be swapped without code changes.
8. Include a safe `bwell-config.yaml` scenario template or equivalent committed example with no literal secrets.
9. Produce a deterministic readiness report with pass/fail status, failure reasons, and redacted evidence for each check.
10. Create a populated `README.md` during implementation using the header-style and tagline pattern from `/Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md`.
11. Keep default tests and normal local development independent of live external credentials.

## NON-GOALS

1. Do not implement a production identity provider or service provider.
2. Do not replace Auth0, b.well, Okta, Entra ID, or any external authorization server.
3. Do not claim b.well account creation, consent, connection search, Smart Connect, or health-record retrieval success unless live b.well responses prove those steps.
4. Do not store, print, snapshot, or commit raw access tokens, ID tokens, refresh tokens, client secrets, client keys, or private credentials by default.
5. Do not hard-code Auth0 or b.well values in application logic.
6. Do not implement SAML in the first release; keep it as a future protocol family.
7. Do not require a database for the first release unless the plan phase proves local ephemeral storage is insufficient.
8. Do not require live b.well credentials for the standard unit test suite.
9. Do not add broad provider SDK abstractions before repeated provider variation proves they are needed.

## USERS

1. Integration engineers validating SSO readiness between an owned application and partner platforms.
2. Application maintainers who own Auth0 or another IdP configuration and need quick feedback before partner onboarding.
3. Partner-integration operators who need a deterministic readiness report showing which identity, token-exchange, or downstream probe step failed.
4. Open-source contributors who want to add new external IdP/SP scenario templates without rewriting core harness behavior.

## SKILLS

Skills are tracked in front matter.

## RELATIONSHIPS

Relationships are tracked in front matter.

## DEPENDENCIES

References are tracked in front matter.

## REQUIREMENTS

### Runtime And Packaging

1. The application must be written in Go.
2. The application must expose a local command-line workflow suitable for open-source users.
3. The application must start a local callback listener for OIDC Authorization Code with PKCE when a scenario requires browser-based login.
4. The application must support a non-live or stubbed path for default verification where external credentials are not available.
5. Source files should remain small and explicit where splitting improves clarity, consistent with `docs/CONSTITUTION.md`.

### Scenario Configuration

1. The application must load one YAML scenario file selected by path.
2. YAML scenario files must describe external provider configuration rather than requiring provider-specific code changes.
3. The scenario schema must include identity-provider settings, callback settings, expected claims, token exchange settings, downstream probe settings, redaction/report settings, and mode selection for stub/live behavior.
4. Unknown or misspelled YAML fields must fail validation before network calls.
5. Required scenario fields must fail validation with actionable errors before network calls.
6. Secret-like values must be represented by environment-variable references or another non-literal secret reference format.
7. The repository must include a safe b.well scenario template named `bwell-config.yaml` or an equivalent clearly documented path.
8. The b.well scenario must model Auth0 as the first IdP and b.well as the first SP without preventing other IdP/SP scenarios.

### OIDC Readiness

1. The harness must discover OIDC provider metadata from the configured issuer.
2. The harness must use PKCE for authorization-code login.
3. The harness must generate and validate state and nonce values.
4. The harness must validate ID token signature through JWKS.
5. The harness must validate issuer, audience, expiration, issued-at tolerance, nonce, and configured expected claims.
6. The harness must summarize token claims in redacted form.
7. The harness must distinguish authentication failures, callback failures, token endpoint failures, discovery failures, JWKS failures, and claim-validation failures in the readiness report.

### Token Exchange And Probes

1. The harness must support generic OAuth 2.0 Token Exchange according to RFC 8693.
2. The harness must support exchanging an IdP ID token as the subject token when configured by scenario.
3. The harness must support b.well's token exchange shape for the first live scenario, including configured b.well client key/header behavior.
4. The harness must support a stub token-exchange mode for deterministic local validation.
5. The harness must support a generic bearer/API probe using the exchanged access token when configured.
6. The harness must classify downstream probe failures separately from authentication and exchange failures.

### Redaction And Reporting

1. Raw tokens and secrets must be redacted by default in logs, terminal output, reports, errors, and test fixtures.
2. Redacted token diagnostics may include token type, selected non-sensitive claims, expiration, issuer, audience, and stable fingerprints that cannot reconstruct the token.
3. The readiness report must be deterministic enough for tests and operator comparison.
4. The readiness report must include per-check status, evidence summary, failure reason, and recommended owner/action when practical.
5. Live b.well sandbox trust failures must be reported as external configuration/trust failures, not as local implementation success.

### Documentation

1. `README.md` must be populated during implementation.
2. The README must use a header-style title block and concise tagline pattern like `/Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md`.
3. The README must include install/build instructions, quick start, scenario YAML usage, b.well scenario setup, redaction guarantees, requirements, and development commands.
4. README commands must match working project commands by the time implementation is complete.

## ACCEPTANCE

1. Running the default test suite does not require live Auth0 or b.well credentials.
2. Scenario YAML validation passes for the committed b.well template when required values are provided through non-secret placeholders or environment references.
3. Scenario YAML validation fails with a clear error for unknown fields, missing required fields, and literal secret values where secret references are required.
4. A stub OIDC/token-exchange flow produces a deterministic readiness report with passing discovery, callback, nonce, issuer, audience, claim, exchange, probe, and redaction checks where applicable.
5. An OIDC callback test proves state mismatch and nonce mismatch are rejected.
6. Token validation tests prove issuer mismatch, audience mismatch, expired token, missing required claim, and invalid signature are rejected.
7. Token exchange tests prove the RFC 8693 request includes the configured grant type, subject token, subject token type, requested token type, and provider-specific headers without logging raw token values.
8. A b.well live-mode run, when credentials and sandbox trust are configured, attempts the configured token exchange and reports the actual b.well response category.
9. A b.well live-mode run, when b.well sandbox trust is incomplete, reports a clear blocked/failing readiness reason and does not claim account/consent/search success.
10. Redaction tests or scans prove raw sample tokens and configured secret values do not appear in normal logs, reports, or test output.
11. README smoke verification confirms documented commands and config paths match the implemented project.
12. `go test ./...` passes after implementation, excluding explicitly documented live-only checks.
13. `go vet ./...` or the repository's chosen equivalent static check passes after implementation.
14. `git diff --check` passes for touched files.

## EDGE-CASES

1. OIDC discovery endpoint unavailable: fail before login with provider-discovery failure and actionable issuer guidance.
2. JWKS unavailable or missing matching key ID: fail token validation and identify key-discovery or signature-verification as the failing check.
3. User cancels login or the IdP returns an authorization error: fail the login step without attempting token exchange.
4. Callback receives missing code, unexpected state, duplicate callback, or nonce mismatch: reject the callback and report callback-validation failure.
5. Token endpoint returns malformed JSON or a non-OAuth error body: preserve redacted response metadata and classify as token-endpoint failure.
6. Expected claim is missing or has the wrong value/type: fail claim validation with the claim name and expected rule, not the raw token.
7. YAML contains unknown fields or conflicting mode settings: fail scenario validation before network calls.
8. YAML references an unset environment variable: fail scenario validation with the variable name but not any secret value.
9. Literal secrets appear in committed scenario config: reject or flag them before running live checks.
10. b.well sandbox trust is not configured: report a blocked/failing live exchange with b.well as the external owner and do not claim downstream readiness.
11. Downstream bearer probe returns 401, 403, 404, 429, or 5xx: classify separately from authentication success and preserve redacted status/body summary.
12. Network timeout or TLS error: fail the affected external check with timeout/TLS classification and continue only where safe.
13. Provider returns a refresh token: redact it everywhere and do not depend on refresh behavior for first-release readiness.
14. Multiple provider scenarios are added later: they must be introduced as scenario YAML plus thin adapter behavior only where standards are insufficient.

## OPEN-QUESTIONS

none
