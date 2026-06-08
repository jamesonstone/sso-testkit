```text
███████╗███████╗ ██████╗
██╔════╝██╔════╝██╔═══██╗
███████╗███████╗██║   ██║
╚════██║╚════██║██║   ██║
███████║███████║╚██████╔╝  testkit
╚══════╝╚══════╝ ╚═════╝
```

**`sso-testkit` is a Go CLI for checking SSO readiness before partner
integration work starts.** It validates scenario configuration, exercises an
OIDC readiness path, checks token exchange shape, probes downstream APIs when
configured, and writes redacted pass/fail evidence.

No raw token dumps. No provider values hard-coded into the app. Just a scenario
file, a readiness run, and a report.

## Install

```sh
git clone https://github.com/jamesonstone/sso-testkit.git
cd sso-testkit
make build
```

## Quick Start

Validate the included token-exchange scenario template:

```sh
make validate
```

Run the local stub readiness path:

```sh
make run
```

Run the test suite:

```sh
make test
make vet
```

## Scenario Files

Scenarios are YAML files. They describe the external identity provider, callback
settings, expected claims, token exchange request, optional bearer probe,
reporting, and redaction policy.

The first template is:

```sh
configs/scenarios/oidc-token-exchange.yaml
```

### Test Pattern File Naming Convention (Safety Guardrail)

To reduce accidental commit risk for live tenant data, use this convention:

- **Commit-safe templates:** `*-template.yaml`
- **Never commit environment/live variants:**
  - `*-local.yaml`
  - `*-dev.yaml`
  - `*-stage.yaml` / `*-staging.yaml`
  - `*-prod.yaml`
  - `*-live.yaml`
  - `*-private.yaml`

Recommended flow:

1. Create and commit a sanitized template (example: `oidc-token-exchange-template.yaml`).
2. Copy it locally to an environment-specific file (example: `oidc-token-exchange-live.yaml`).
3. Put real tenant values only in the local/live file.
4. Keep secrets in env refs (`env:...`) and runtime environment variables.

This repository's `.gitignore` includes patterns for the environment/live suffixes
above to help prevent accidental publication of real test data.

### Canonical Scenario Documentation File Protection

`configs/scenarios/oidc-token-exchange.yaml` is intentionally kept in-repo as
canonical documentation.

To avoid accidental follow-up commits to that file (for example, adding real
tenant endpoints or secrets), this repo includes a versioned pre-commit hook
that warns when that specific file is staged.

Enable the guard once per clone:

```sh
make hooks-install
```

If you intentionally need to update the documentation file, bypass the guard
is not required in default mode (warning-only).

If you want stricter local safety, enable strict mode per commit:

```sh
STRICT_SCENARIO_DOC_GUARD=1 git commit ...
```

In strict mode, bypass once for intentional documentation updates:

```sh
ALLOW_SCENARIO_DOC_UPDATE=1 git commit ...
```

It models a generic OIDC identity provider and a generic token-exchange service.
Provider-specific values stay in YAML so another IdP or downstream service can be
plugged in without changing core code.

Secret-like fields use environment references:

```yaml
headers:
  X-API-Key: env:PARTNER_API_KEY
```

Unset live-mode environment references fail validation before network calls.
Stub mode does not require live external credentials.

## Live Mode

Live readiness is opt-in and depends on external trust and credentials being
configured outside this repository.

Use an ignored environment-specific file for live runs (for example,
`configs/scenarios/oidc-token-exchange-live.yaml`) rather than committing
real values into a shared template file.

```sh
export PARTNER_API_KEY='...'
go run ./cmd/sso-testkit run --config configs/scenarios/oidc-token-exchange-live.yaml --mode live --report readiness-report.json
```

If external trust is incomplete, the report classifies the exchange as an
external trust or configuration failure. The tool only reports checks that are
actually exercised by the configured live scenario.

## Redaction

`sso-testkit` redacts raw tokens and secrets by default in reports, errors, and
diagnostic output. Reports may include token types, selected non-sensitive
claims, status codes, and non-reversible fingerprints for correlation.

## Requirements

- Go 1.25+
- Network access for live OIDC and downstream checks
- External provider credentials only for live mode

## Development

```sh
make fmt
make vet
make test
make build
git diff --check
make validate
make run
```

## Maintenance

Made with ❤️ by @jamesonstone. Contributions welcome via pull request or issue.
