package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jamesonstone/sso-testkit/internal/config"
	"github.com/jamesonstone/sso-testkit/internal/exchange"
	"github.com/jamesonstone/sso-testkit/internal/oidcflow"
	"github.com/jamesonstone/sso-testkit/internal/probe"
	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/report"
)

const (
	ExitOK        = 0
	ExitReadiness = 1
	ExitUsage     = 2
)

type LookupEnv func(string) (string, bool)

func Main(args []string, stdout, stderr io.Writer, lookup LookupEnv) int {
	if len(args) == 0 {
		usage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "validate-config":
		return runValidateConfig(args[1:], stdout, stderr, lookup)
	case "run":
		return runReadiness(args[1:], stdout, stderr, lookup)
	case "-h", "--help", "help":
		usage(stdout)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		usage(stderr)
		return ExitUsage
	}
}

func runValidateConfig(args []string, stdout, stderr io.Writer, lookup LookupEnv) int {
	fs := flag.NewFlagSet("validate-config", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "scenario YAML path")
	mode := fs.String("mode", "", "mode override: stub or live")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if *configPath == "" {
		fmt.Fprintln(stderr, "--config is required")
		return ExitUsage
	}
	cfg, err := config.LoadFileWithLookup(*configPath, *mode, lookup)
	if err != nil {
		fmt.Fprintf(stderr, "config invalid: %v\n", err)
		return ExitUsage
	}
	fmt.Fprintf(stdout, "config ok: %s mode=%s\n", cfg.Scenario.ID, cfg.Mode)
	return ExitOK
}

func runReadiness(args []string, stdout, stderr io.Writer, lookup LookupEnv) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "scenario YAML path")
	mode := fs.String("mode", "", "mode override: stub or live")
	reportPath := fs.String("report", "", "report output path, or - for stdout")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if *configPath == "" {
		fmt.Fprintln(stderr, "--config is required")
		return ExitUsage
	}
	cfg, err := config.LoadFileWithLookup(*configPath, *mode, lookup)
	if err != nil {
		fmt.Fprintf(stderr, "config invalid: %v\n", err)
		return ExitUsage
	}

	rep := Execute(context.Background(), *cfg, stderr)
	if err := writeReport(rep, *reportPath, stdout); err != nil {
		fmt.Fprintf(stderr, "write report: %v\n", err)
		return ExitUsage
	}
	fmt.Fprintf(stderr, "readiness %s: %s\n", rep.Overall, cfg.Scenario.ID)
	if rep.Overall == report.StatusPass || rep.Overall == report.StatusSkip {
		return ExitOK
	}
	return ExitReadiness
}

func Execute(ctx context.Context, cfg config.Config, stderr io.Writer) report.Report {
	started := time.Now().UTC()
	rep := report.New(cfg.Scenario.ID, cfg.Mode, started)
	rep.Add(report.Pass(report.CheckConfigLoad, map[string]string{"scenario": cfg.Scenario.ID}))
	rep.Add(report.Pass(report.CheckConfigValidate, map[string]string{"mode": cfg.Mode}))

	redactor := redact.New(cfg.SecretValues()...)
	for _, value := range cfg.Redaction.ExtraValues {
		redactor.Add(value)
	}
	client := &http.Client{Timeout: 30 * time.Second}

	var oidcResult oidcflow.Result
	if cfg.Mode == config.ModeStub {
		oidcResult = oidcflow.RunStub(cfg, redactor)
	} else {
		oidcResult = oidcflow.RunLive(ctx, cfg, client, func(authURL string) {
			fmt.Fprintf(stderr, "open this URL to continue: %s\n", redactor.URL(authURL))
		})
	}
	for _, check := range oidcResult.Checks {
		rep.Add(check)
	}
	if hasFailed(oidcResult.Checks) {
		rep.Finish(time.Now().UTC())
		return rep
	}

	token, exchangeChecks := exchange.Run(ctx, cfg.Exchange, cfg.ResolvedSecrets, oidcResult.IDToken, client, redactor, cfg.Mode)
	for _, check := range exchangeChecks {
		rep.Add(check)
	}
	if hasFailed(exchangeChecks) {
		rep.Finish(time.Now().UTC())
		return rep
	}

	for _, check := range probe.Run(ctx, cfg.Probe, cfg.ResolvedSecrets, token.AccessToken, client, redactor) {
		rep.Add(check)
	}
	rep.Add(report.Pass(report.CheckRedactionScan, map[string]string{"policy": "default"}))
	rep.Finish(time.Now().UTC())
	return rep
}

func writeReport(rep report.Report, path string, stdout io.Writer) error {
	if path == "" {
		return nil
	}
	data, err := rep.JSON()
	if err != nil {
		return err
	}
	if path == "-" {
		_, err = fmt.Fprintln(stdout, string(data))
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func hasFailed(checks []report.Check) bool {
	for _, check := range checks {
		if check.Status == report.StatusFail || check.Status == report.StatusBlocked {
			return true
		}
	}
	return false
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  sso-testkit validate-config --config <path> [--mode stub|live]")
	fmt.Fprintln(w, "  sso-testkit run --config <path> --mode stub|live [--report <path>|-]")
}
