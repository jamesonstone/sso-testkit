package secretref

import (
	"fmt"
	"regexp"
	"strings"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Ref struct {
	Raw string
	Env string
}

func IsRef(value string) bool {
	return strings.HasPrefix(value, "env:")
}

func Parse(value string) (Ref, error) {
	if !IsRef(value) {
		return Ref{}, fmt.Errorf("secret reference must use env:VARIABLE_NAME")
	}
	env := strings.TrimPrefix(value, "env:")
	if !envNamePattern.MatchString(env) {
		return Ref{}, fmt.Errorf("invalid environment variable name %q", env)
	}
	return Ref{Raw: value, Env: env}, nil
}

func Resolve(value string, lookup func(string) (string, bool)) (string, error) {
	ref, err := Parse(value)
	if err != nil {
		return "", err
	}
	secret, ok := lookup(ref.Env)
	if !ok || secret == "" {
		return "", fmt.Errorf("environment variable %s is not set", ref.Env)
	}
	return secret, nil
}
