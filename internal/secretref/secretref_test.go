package secretref

import "testing"

func TestParse(t *testing.T) {
	ref, err := Parse("env:PARTNER_API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Env != "PARTNER_API_KEY" {
		t.Fatalf("Env = %q", ref.Env)
	}
}

func TestParseRejectsLiteral(t *testing.T) {
	if _, err := Parse("secret-value"); err == nil {
		t.Fatal("expected literal secret to be rejected")
	}
}

func TestResolveReportsVariableNameOnly(t *testing.T) {
	_, err := Resolve("env:MISSING_SECRET", func(string) (string, bool) { return "", false })
	if err == nil {
		t.Fatal("expected missing env var error")
	}
	if got := err.Error(); got != "environment variable MISSING_SECRET is not set" {
		t.Fatalf("unexpected error: %s", got)
	}
}
