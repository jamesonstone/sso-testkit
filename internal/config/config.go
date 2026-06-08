package config

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jamesonstone/sso-testkit/internal/redact"
	"github.com/jamesonstone/sso-testkit/internal/secretref"
	"gopkg.in/yaml.v3"
)

const (
	ModeStub = "stub"
	ModeLive = "live"
)

type Config struct {
	Scenario       Scenario     `yaml:"scenario"`
	ModeDefaults   ModeDefaults `yaml:"mode_defaults"`
	Callback       Callback     `yaml:"callback"`
	IDP            IDP          `yaml:"idp"`
	ExpectedClaims []ClaimRule  `yaml:"expected_claims"`
	Exchange       Exchange     `yaml:"exchange"`
	Probe          Probe        `yaml:"probe"`
	Report         Report       `yaml:"report"`
	Redaction      Redaction    `yaml:"redaction"`

	Mode            string            `yaml:"-"`
	ResolvedSecrets map[string]string `yaml:"-"`
}

type Scenario struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Providers   Providers `yaml:"providers"`
}

type Providers struct {
	IDP string `yaml:"idp"`
	SP  string `yaml:"sp"`
}

type ModeDefaults struct {
	Mode string `yaml:"mode"`
}

type Callback struct {
	Bind        string `yaml:"bind"`
	Port        int    `yaml:"port"`
	Path        string `yaml:"path"`
	RedirectURI string `yaml:"redirect_uri"`
	Timeout     string `yaml:"timeout"`
}

func (c Callback) TimeoutDuration() time.Duration {
	if c.Timeout == "" {
		return 5 * time.Minute
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 0
	}
	return d
}

type IDP struct {
	Issuer       string   `yaml:"issuer"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	Scopes       []string `yaml:"scopes"`
	RedirectURI  string   `yaml:"redirect_uri"`
	Audience     string   `yaml:"audience"`
	ExpectedAlgs []string `yaml:"expected_algs"`
}

type ClaimRule struct {
	Name          string   `yaml:"name"`
	Required      bool     `yaml:"required"`
	ExpectedValue string   `yaml:"expected_value"`
	AllowedValues []string `yaml:"allowed_values"`
	Redact        bool     `yaml:"redact"`
}

type Exchange struct {
	Profile            string            `yaml:"profile"`
	Endpoint           string            `yaml:"endpoint"`
	Method             string            `yaml:"method"`
	Headers            map[string]string `yaml:"headers"`
	SubjectTokenSource string            `yaml:"subject_token_source"`
	SubjectTokenType   string            `yaml:"subject_token_type"`
	RequestedTokenType string            `yaml:"requested_token_type"`
	Resource           string            `yaml:"resource"`
	Audience           string            `yaml:"audience"`
	Scope              string            `yaml:"scope"`
	Timeout            string            `yaml:"timeout"`
	StubResponse       StubTokenResponse `yaml:"stub_response"`
}

func (e Exchange) TimeoutDuration() time.Duration {
	if e.Timeout == "" {
		return 10 * time.Second
	}
	d, err := time.ParseDuration(e.Timeout)
	if err != nil {
		return 0
	}
	return d
}

type StubTokenResponse struct {
	AccessToken     string `yaml:"access_token"`
	IssuedTokenType string `yaml:"issued_token_type"`
	TokenType       string `yaml:"token_type"`
	ExpiresIn       int    `yaml:"expires_in"`
}

type Probe struct {
	Enabled        bool              `yaml:"enabled"`
	URL            string            `yaml:"url"`
	Method         string            `yaml:"method"`
	Headers        map[string]string `yaml:"headers"`
	ExpectedStatus []int             `yaml:"expected_status"`
	Timeout        string            `yaml:"timeout"`
	RedactBody     bool              `yaml:"redact_body"`
}

func (p Probe) TimeoutDuration() time.Duration {
	if p.Timeout == "" {
		return 10 * time.Second
	}
	d, err := time.ParseDuration(p.Timeout)
	if err != nil {
		return 0
	}
	return d
}

type Report struct {
	Output string `yaml:"output"`
}

type Redaction struct {
	ExtraValues []string `yaml:"extra_values"`
}

func LoadFile(path, modeOverride string) (*Config, error) {
	return LoadFileWithLookup(path, modeOverride, os.LookupEnv)
}

func LoadFileWithLookup(path, modeOverride string, lookup func(string) (string, bool)) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg, err := Parse(data)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(modeOverride, lookup); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Parse(data []byte) (*Config, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Validate(modeOverride string, lookup func(string) (string, bool)) error {
	var errs []error
	c.ResolvedSecrets = map[string]string{}
	mode := modeOverride
	if mode == "" {
		mode = c.ModeDefaults.Mode
	}
	if mode == "" {
		mode = ModeStub
	}
	if mode != ModeStub && mode != ModeLive {
		errs = append(errs, fmt.Errorf("mode must be %q or %q", ModeStub, ModeLive))
	}
	c.Mode = mode

	required := map[string]string{
		"scenario.id":                   c.Scenario.ID,
		"idp.issuer":                    c.IDP.Issuer,
		"idp.client_id":                 c.IDP.ClientID,
		"exchange.profile":              c.Exchange.Profile,
		"exchange.endpoint":             c.Exchange.Endpoint,
		"exchange.subject_token_type":   c.Exchange.SubjectTokenType,
		"exchange.requested_token_type": c.Exchange.RequestedTokenType,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Errorf("%s is required", field))
		}
	}

	if len(c.IDP.Scopes) == 0 {
		errs = append(errs, errors.New("idp.scopes requires at least one scope"))
	}
	if !contains(c.IDP.Scopes, "openid") {
		errs = append(errs, errors.New("idp.scopes must include openid"))
	}

	errs = append(errs, validateURL("idp.issuer", c.IDP.Issuer)...)
	errs = append(errs, validateURL("exchange.endpoint", c.Exchange.Endpoint)...)
	if c.Callback.RedirectURI != "" {
		errs = append(errs, validateURL("callback.redirect_uri", c.Callback.RedirectURI)...)
	}
	if c.IDP.RedirectURI != "" {
		errs = append(errs, validateURL("idp.redirect_uri", c.IDP.RedirectURI)...)
	}
	if c.RedirectURI() == "" {
		errs = append(errs, errors.New("callback.redirect_uri or idp.redirect_uri is required"))
	}
	if c.Callback.Path != "" && !strings.HasPrefix(c.Callback.Path, "/") {
		errs = append(errs, errors.New("callback.path must start with /"))
	}
	if c.Callback.Port < 0 || c.Callback.Port > 65535 {
		errs = append(errs, errors.New("callback.port must be between 0 and 65535"))
	}
	if c.Callback.TimeoutDuration() <= 0 {
		errs = append(errs, errors.New("callback.timeout must be a positive duration"))
	}
	if c.Exchange.TimeoutDuration() <= 0 {
		errs = append(errs, errors.New("exchange.timeout must be a positive duration"))
	}
	if c.Probe.Enabled {
		errs = append(errs, validateURL("probe.url", c.Probe.URL)...)
		if c.Probe.TimeoutDuration() <= 0 {
			errs = append(errs, errors.New("probe.timeout must be a positive duration"))
		}
	}

	for i, claim := range c.ExpectedClaims {
		if strings.TrimSpace(claim.Name) == "" {
			errs = append(errs, fmt.Errorf("expected_claims[%d].name is required", i))
		}
	}

	if c.IDP.ClientSecret != "" {
		errs = append(errs, c.validateSecret("idp.client_secret", c.IDP.ClientSecret, mode == ModeLive, lookup)...)
	}
	for key, value := range c.Exchange.Headers {
		field := "exchange.headers." + key
		if redact.SensitiveName(key) {
			errs = append(errs, c.validateSecret(field, value, mode == ModeLive, lookup)...)
		}
	}
	for key, value := range c.Probe.Headers {
		field := "probe.headers." + key
		if redact.SensitiveName(key) {
			errs = append(errs, c.validateSecret(field, value, mode == ModeLive, lookup)...)
		}
	}
	if mode == ModeStub && c.Exchange.StubResponse.AccessToken == "" {
		errs = append(errs, errors.New("exchange.stub_response.access_token is required for stub mode"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c *Config) validateSecret(field, value string, resolve bool, lookup func(string) (string, bool)) []error {
	if !secretref.IsRef(value) {
		return []error{fmt.Errorf("%s must use env:VARIABLE_NAME", field)}
	}
	if _, err := secretref.Parse(value); err != nil {
		return []error{fmt.Errorf("%s: %w", field, err)}
	}
	if !resolve {
		return nil
	}
	secret, err := secretref.Resolve(value, lookup)
	if err != nil {
		return []error{fmt.Errorf("%s: %w", field, err)}
	}
	c.ResolvedSecrets[field] = secret
	return nil
}

func (c Config) SecretValues() []string {
	values := make([]string, 0, len(c.ResolvedSecrets)+len(c.Redaction.ExtraValues))
	for _, value := range c.ResolvedSecrets {
		values = append(values, value)
	}
	values = append(values, c.Redaction.ExtraValues...)
	sort.Strings(values)
	return values
}

func (c Config) RedirectURI() string {
	if c.IDP.RedirectURI != "" {
		return c.IDP.RedirectURI
	}
	return c.Callback.RedirectURI
}

func validateURL(field, value string) []error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return []error{fmt.Errorf("%s must be an absolute URL", field)}
	}
	return nil
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
