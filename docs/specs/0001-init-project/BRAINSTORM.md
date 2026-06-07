---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0001-init-project
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: project constraints and workflow classification
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
    used_for: documentation completion bar and safety rules
    status: active
  - id: agent-rlm
    name: RLM context routing
    type: repo_doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: conditional
    used_for: just-in-time prior-work and repository discovery
    status: active
  - id: kit-map-init-project
    name: kit map 0001-init-project
    type: command
    target: "kit map 0001-init-project"
    selector_type: command
    selector: "kit map 0001-init-project"
    relation: informs
    read_policy: evidence
    used_for: confirming current feature phase and absence of prior relationships
    status: active
  - id: project-progress
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: informs
    read_policy: conditional
    used_for: confirming this is the only tracked feature
    status: active
  - id: yp-readme
    name: yp README
    type: local_reference
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md
    relation: informs
    read_policy: evidence
    used_for: README header-style and tagline pattern requested by user
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
relationships: []
---
# BRAINSTORM

## SUMMARY

Build `sso-testkit` as a small Go-based SSO readiness harness that validates an owned application against external identity-provider and service-provider configurations before production integration work starts. The first implementation direction is a protocol-first OIDC Authorization Code with PKCE flow, redacted diagnostics, YAML-driven test scenarios, and an Auth0-to-b.well preset that remains reusable for other providers.

## USER THESIS

The user wants `sso-testkit` to become a reusable, open-source SSO readiness harness rather than a one-off b.well script. The harness should be implemented in Go, use YAML scenario files to plug in external IdP/SP configurations, keep token and secret diagnostics redacted, and make Auth0-to-b.well the first scenario without hard-coding that pair into the core product.

## Context Synthesis

Objective: create a small open-source SSO readiness harness that validates an owned client system against an identity provider, inspects redacted tokens or assertions, and probes a downstream system through token exchange or bearer/API calls before production integration work starts [S1][S2]. Affected users are integration engineers configuring SSO between owned applications and partner platforms, with b.well serving as the first preset rather than the core architecture [S1][S2][S3].

The target implementation language is now resolved: this application must be written in Go and should use idiomatic Go structure, explicit error handling, focused packages, small interfaces only at real boundaries, and standard-library primitives unless a mature dependency is warranted for OIDC, OAuth2, JWT validation, or YAML parsing [S7].

Configuration should be scenario-driven. Because the application is the test harness and not itself the external IdP or SP, YAML scenario files are required to decouple external provider configuration from code. The initial scenario should support an internal Auth0 connection as IdP against b.well as SP through a `bwell-config.yaml` scenario file, with secrets supplied through environment variables or secret references rather than committed literal values [S9].

Constraints: secrets and raw tokens stay redacted by default, local callback flow is the default runtime, provider-specific behavior stays in adapters, README documentation must be useful at project initialization, and live b.well acceptance cannot pass until b.well completes sandbox trust configuration [S1][S2][S3][S8]. Definition of done: OIDC Authorization Code with PKCE works, OIDC discovery and JWKS validation work, expected claims are checked, RFC 8693 token exchange works in stub and live modes, generic bearer/API probe works, YAML scenarios select provider configuration without code changes, and a deterministic readiness report identifies pass/fail reasons [S2][S3][S6][S7][S9]. Selected direction: build a protocol-first Go harness with Auth0-to-b.well as the first configuration preset [S2][S3][S7].

## Source Map

- [S1] discussion: User is waiting for b.well sandbox configuration and wants a small app to test SSO readiness. Source: conversation:side-thread:bwell-sandbox-readiness.
- [S2] discussion: User wants the tool to become a reusable open-source project for SSO configurations where one system is owned. Source: conversation:side-thread:generic-oss-sso-harness.
- [S3] link: b.well supports OAuth Token Exchange with an ID token as subject token. Source: https://developer.bwell.com/docs/oauth-token-exchange
- [S4] link: b.well connection search requires end-user authentication, account creation, and consent. Source: https://developer.bwell.com/docs/search-establish-data-connections
- [S5] link: b.well Smart Connect requires authenticated users, account creation, consent, and identity verification prerequisites. Source: https://developer.bwell.com/docs/smart-connect-locating-managing-connections
- [S6] link: OAuth 2.0 Token Exchange is standardized by RFC 8693. Source: https://www.rfc-editor.org/rfc/rfc8693
- [S7] discussion: User requires the application to be written in Golang and to leverage best practices. Source: current request.
- [S8] local reference: README should follow the header-style and tagline pattern used by `yp`. Source: /Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md.
- [S9] discussion: User requires YAML configuration files per test scenario for external providers when the app is not itself the external provider; example scenario is internal Auth0 IdP against b.well SP. Source: current request.

## Coding Agent Instructions

Chosen direction: implement or specify a reusable Go SSO readiness harness with protocol modules, redacted diagnostics, YAML scenario configuration, configurable downstream probes, and a b.well preset; tradeoff: keep provider presets thin so the first release has less b.well-specific automation and stronger reuse across Auth0, Okta, Entra ID, and partner APIs [S2][S3][S6][S7][S9]. The target repository is `/Users/jamesonstone/go/src/github.com/jamesonstone/sso-testkit`; current inspection shows a documentation-only scaffold, so the next spec and plan phases must create the Go application surface from scratch [S2][S7].

1. Inspect the repository and identify existing functionality by exact file path and symbol: run `pwd`, `git status --short`, `rg --files`, inspect package manifests, app entrypoints, auth modules, server/router files, test setup, env examples, and README files [S2].
2. Reconcile brainstorm decisions with actual code behavior; write `CONFLICT` for stale docs or code mismatch, and use verified current code plus user decisions as the tie-break decision [S1][S2].
3. Implement the application in Go. Prefer standard library `net/http`, `context`, `crypto/rand`, `encoding/json`, `log/slog`, and `flag` where sufficient; use mature dependencies for OIDC/OAuth2/JWT/YAML only where they remove risky security-sensitive code [S7].
4. Produce a complete implementation strategy grounded in current codebase context, including Go module layout, binary shape, CLI/server shape, local callback URL, session storage, redaction policy, scenario YAML schema, and report format [S2][S6][S7][S9].
5. Enumerate concrete file edits, interfaces, data model changes, dependency updates, configuration changes, migration steps, validation commands, and tests before editing code [S2].
6. Implement OIDC Authorization Code with PKCE, OIDC discovery, JWKS validation, state and nonce validation, expected-claim checks, and redacted token inspection [S2].
7. Implement downstream adapters for generic RFC 8693 token exchange, generic bearer/API probe, stub b.well token exchange, and live b.well token exchange using `clientkey` header configuration [S3][S6].
8. Add a b.well scenario/preset that exchanges an Auth0 ID token, reports sandbox trust failures clearly, and blocks account/consent/search success claims until live b.well responses pass [S1][S3][S4][S9].
9. Add scenario YAML support before hard-coding provider behavior. The first checked-in scenario should be `configs/scenarios/bwell-config.yaml` or an equivalent safe template with no literal secrets; the binary should accept a config path so a separate provider can be plugged in without code changes [S9].
10. Update `README.md` during implementation so it is populated at project birth. It should use a header-style block and tagline pattern like `/Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md`, then include install, quick start, scenario configuration, b.well example, redaction guarantees, requirements, and development commands [S8].
11. Define acceptance checks with expected command outputs: login callback pass, nonce pass, issuer/audience pass, scenario YAML load pass, stub exchange pass, live exchange returns configured b.well result, README smoke instructions match real commands, and redaction scan shows no raw token output [S2][S3][S8][S9].
12. State risks, open questions, explicit assumptions, mitigation, and owner: b.well sandbox trust configuration owner is b.well; client credentials and callback configuration owner is repository maintainer; token logging mitigation is default redaction plus tests [S1][S2][S3].

## Resource Links

- [R1] b.well Developer Portal — https://developer.bwell.com/ — Main b.well developer documentation entrypoint.
- [R2] b.well Welcome — https://developer.bwell.com/docs/welcome — Orientation for b.well platform documentation.
- [R3] b.well Auth Overview — https://developer.bwell.com/docs/auth-overview — Authentication overview for b.well integrations.
- [R4] b.well End-User Auth — https://developer.bwell.com/docs/end-user-auth — End-user authentication reference for patient-scoped flows.
- [R5] b.well OAuth Token Exchange — https://developer.bwell.com/docs/oauth-token-exchange — First b.well preset token exchange target.
- [R6] b.well Account Creation and Consent — https://developer.bwell.com/docs/account-creation-consent-1 — Account creation and consent workflow reference.
- [R7] b.well Search & Establish Data Connections — https://developer.bwell.com/docs/search-establish-data-connections — Downstream provider-search workflow and prerequisites.
- [R8] b.well Smart Connect — https://developer.bwell.com/docs/smart-connect-locating-managing-connections — Smart Connect prerequisite and connection workflow reference.
- [R9] b.well Data Exports API — https://developer.bwell.com/reference/post_users-id-data-exports — Related b.well user data export endpoint.
- [R10] RFC 8693 OAuth 2.0 Token Exchange — https://www.rfc-editor.org/rfc/rfc8693 — Generic downstream token exchange standard.
- [R11] RFC 6749 OAuth 2.0 — https://www.rfc-editor.org/rfc/rfc6749 — Authorization framework baseline.
- [R12] RFC 7636 PKCE — https://www.rfc-editor.org/rfc/rfc7636 — PKCE standard for public-client authorization code flow.
- [R13] OpenID Connect Core 1.0 — https://openid.net/specs/openid-connect-core-1_0.html — ID token, nonce, issuer, subject, and audience validation rules.
- [R14] OpenID Connect Discovery 1.0 — https://openid.net/specs/openid-connect-discovery-1_0.html — Provider metadata discovery standard.
- [R15] RFC 7517 JSON Web Key — https://www.rfc-editor.org/rfc/rfc7517 — JWKS validation reference.
- [R16] Auth0 OpenID Connect Protocol — https://auth0.com/docs/authenticate/protocols/openid-connect-protocol — Auth0 OIDC behavior reference for the first IdP scenario.
- [R17] OASIS SAML Core 2.0 — https://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf — Future SAML assertion support reference.
- [R18] yp README — /Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md — README header-style, tagline, quick start, requirements, and development section pattern.

## RELATIONSHIPS

No prior feature relationships. Front matter records `relationships: []`, and `kit map 0001-init-project` reports no incoming or outgoing relationships.

## CODEBASE FINDINGS

1. Current repository state is a Kit scaffold, not an implemented application. `rg --files --hidden -g '!.git/**'` shows repo instructions, `.kit.yaml`, `.gitignore`, `.coderabbit.yaml`, `.github/*`, `docs/*`, and `docs/specs/0001-init-project/BRAINSTORM.md`, but no `README.md`, `go.mod`, `cmd/`, `internal/`, test files, runtime config, or application entrypoint.
2. `docs/specs/0001-init-project/BRAINSTORM.md` is the only current feature artifact. `kit map 0001-init-project` reports phase `brainstorm`, with `SPEC.md`, `PLAN.md`, and `TASKS.md` missing.
3. `docs/notes/0001-init-project` contains only `.gitkeep`; there are no feature-note inputs to preserve or promote into the brainstorm.
4. `.kit.yaml` sets `goal_percentage: 95` and `loop.min_confidence: 95`, matching the research prompt's confidence bar.
5. `.gitignore` is already Go-oriented: it excludes binaries, test binaries, coverage profiles, `go.work`, `.env`, and Kit scratch/cache state. That supports a Go implementation but may need future additions for local scenario files if real provider YAML files contain environment-specific values.
6. `docs/CONSTITUTION.md` is mostly baseline Kit guidance. The useful invariant for this feature is that implementation/source files should stay around 300 lines when splitting improves clarity, while docs are exempt from that guideline.
7. `docs/references/testing.md`, `docs/references/tooling.md`, and `docs/references/external-systems.md` contain only placeholder repo-wide reference guidance. Feature-specific protocol, provider, config, and testing detail should stay in this feature's docs until stable enough to promote.
8. `/Users/jamesonstone/go/src/github.com/jamesonstone/yp/README.md` starts with a header-style text block, a bold tagline paragraph, a short no-friction positioning sentence, then `Install`, `Quick Start`, `Behavior`, `Requirements`, and `Development`. The `sso-testkit` README should use that pattern without copying `yp`-specific behavior.
9. The b.well OAuth Token Exchange page shows the first b.well scenario should exchange an IdP ID token for a b.well access token using OAuth 2.0 Token Exchange, then use the returned b.well token for b.well API calls.
10. OpenID Connect Discovery requires provider metadata such as issuer, authorization endpoint, token endpoint, and JWKS URI; this supports using discovery and JWKS validation instead of hand-maintained endpoint lists.
11. RFC 7636 documents PKCE as a mitigation for authorization code interception in public clients; the local callback harness should use PKCE by default.
12. RFC 8693 defines token exchange request/response semantics and token-type identifiers; the generic exchange adapter should implement the standard first and keep b.well-specific request details thin.

## AFFECTED FILES

1. `docs/specs/0001-init-project/BRAINSTORM.md` - current research artifact and source of this update.
2. `README.md` - absent today; future implementation must create a populated README using the `yp` header-style and tagline pattern, with real commands and configuration examples.
3. `go.mod` - absent today; future implementation must create the Go module and pin mature dependencies intentionally.
4. `cmd/sso-testkit/main.go` - proposed future binary entrypoint for CLI orchestration and local callback server startup.
5. `internal/config/` - proposed future package for strict YAML scenario loading, environment interpolation, validation, and redaction metadata.
6. `configs/scenarios/bwell-config.yaml` - proposed first scenario file or safe template for internal Auth0 IdP to b.well SP readiness checks; committed content must not contain literal secrets.
7. `internal/oidc/` - proposed future package for OIDC discovery, PKCE, callback state, nonce validation, and ID token verification.
8. `internal/exchange/` - proposed future package for generic RFC 8693 token exchange plus b.well-specific header/request configuration.
9. `internal/probe/` - proposed future package for downstream bearer/API probes.
10. `internal/report/` - proposed future package for deterministic pass/fail readiness report generation and redacted diagnostic output.
11. `.gitignore` - may need future adjustment if real local scenario files, output reports, or local callback artifacts should be ignored.
12. `docs/specs/0001-init-project/SPEC.md`, `docs/specs/0001-init-project/PLAN.md`, and `docs/specs/0001-init-project/TASKS.md` - missing today; the next Kit phases should turn this brainstorm into requirements, implementation strategy, and executable task sequencing.

## DEPENDENCIES

1. Required runtime language: Go. The implementation should use idiomatic Go package boundaries and avoid non-Go application frameworks.
2. Likely Go dependencies:
   - `golang.org/x/oauth2` for OAuth2 Authorization Code flow mechanics.
   - `github.com/coreos/go-oidc/v3/oidc` or an equivalent mature OIDC validator for discovery, JWKS-backed ID token verification, and claim validation.
   - `gopkg.in/yaml.v3` or an equivalent YAML library with strict known-field decoding for scenario config.
3. Default standard-library dependencies:
   - `net/http` for local callback and provider/probe calls.
   - `context` for cancellation and request scoping.
   - `crypto/rand` and `crypto/sha256` for state, nonce, PKCE verifier/challenge generation.
   - `encoding/json` for report output.
   - `log/slog` for structured redacted diagnostics.
   - `flag` for a minimal CLI unless later requirements justify a CLI framework.
4. External protocol/provider dependencies:
   - OIDC Core and Discovery for authorization, claims, issuer/audience/nonce validation, and JWKS metadata.
   - RFC 7636 for PKCE.
   - RFC 7517 for JWKS.
   - RFC 8693 for token exchange.
   - b.well OAuth Token Exchange documentation for the first service-provider scenario.
   - Auth0 OIDC documentation for the first identity-provider scenario.
5. Secret handling dependency: scenario YAML should use environment-variable references for client secret, b.well client key, and other credentials. Real token values and secrets must never be committed or printed by default.

## QUESTIONS

No blocking questions remain for adding the three new requirements to this brainstorm. Current understanding is 95%.

Recommended defaults carried forward unless the user overrides them:

1. Go implementation is mandatory; use standard-library components first and mature security/protocol libraries where hand-rolling would create unnecessary risk.
2. The application is a harness that talks to external providers, not the external provider itself, so per-scenario YAML is required.
3. The first scenario should be an Auth0-to-b.well config named `bwell-config.yaml` under a scenario config directory.
4. Committed YAML must avoid literal secrets; secrets should come from environment variables or secret references resolved at runtime.
5. README creation is part of the initial implementation, not a later polish task.

## OPTIONS

1. Go protocol-first harness with YAML scenarios. This is the recommended option. It satisfies the user's language requirement, keeps provider-specific data out of code, supports Auth0-to-b.well first, and leaves room for Okta, Entra ID, or other service providers without a rewrite.
2. Go app with hard-coded b.well/Auth0 values. This is faster for a one-off test, but it fails the plug-in provider requirement and would make the open-source project less reusable.
3. App-as-external-provider design. This would omit external-provider YAML, but it does not match the current thesis: the tool should validate an owned client system against external IdP/SP behavior, not become the IdP or SP under test.
4. Non-Go implementation. This is rejected by the user's current requirement.
5. Full provider SDK abstraction up front. This may become useful later, but it is premature for the first release. Thin adapters around protocol behavior are enough until repeated provider-specific variation appears.

## RECOMMENDED STRATEGY

1. In `SPEC.md`, lock the product as a Go command-line harness that starts a local callback server for OIDC Authorization Code with PKCE, validates ID tokens through OIDC Discovery and JWKS, applies configured claim expectations, exchanges tokens or calls downstream probes, and writes a deterministic readiness report.
2. Define a strict YAML scenario schema before implementing provider logic. The schema should include `idp`, `callback`, `expected_claims`, `exchange`, `probe`, `redaction`, and `report` sections. Unknown fields should fail fast so misspelled provider settings do not silently produce misleading readiness results.
3. Add `configs/scenarios/bwell-config.yaml` as the first safe scenario template. It should model internal Auth0 as IdP and b.well as SP, use environment references for secrets, and keep b.well-specific fields limited to exchange endpoint, `clientkey` source, token-type settings, expected response behavior, and optional probe configuration.
4. Keep adapters thin. Build generic OIDC, token exchange, bearer probe, report, and redaction components first; put b.well behavior behind configuration and a small adapter only where b.well differs from the generic standard.
5. Create `README.md` during initial implementation. Start with a header-style text block and concise tagline like `yp`, then provide install, quick start, b.well scenario setup, YAML schema example, redaction behavior, requirements, and development commands.
6. Validate with unit tests for YAML parsing, redaction, PKCE/state/nonce helpers, report rendering, and token exchange request construction; add integration-style stub tests for the local callback and stub exchange path. Live b.well checks should be explicit/manual or gated by environment so default tests do not require external credentials.
7. Treat raw tokens and secrets as toxic data. Default output should show only redacted token fingerprints, claim summaries, issuer/audience/nonce status, exchange/probe status, and actionable failure reasons.

## NEXT STEP

Run `kit spec init-project` to turn this brainstorm into a formal specification, then `kit plan init-project` to produce the implementation plan before any product code is written.
