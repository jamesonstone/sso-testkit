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

```sh
export PARTNER_API_KEY='...'
go run ./cmd/sso-testkit run --config configs/scenarios/oidc-token-exchange.yaml --mode live --report readiness-report.json
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
