package redact

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var jwtPattern = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)

type Redactor struct {
	values []string
}

func New(values ...string) Redactor {
	r := Redactor{}
	for _, value := range values {
		r.Add(value)
	}
	return r
}

func (r *Redactor) Add(value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	for _, existing := range r.values {
		if existing == value {
			return
		}
	}
	r.values = append(r.values, value)
	sort.Slice(r.values, func(i, j int) bool {
		return len(r.values[i]) > len(r.values[j])
	})
}

func (r Redactor) String(value string) string {
	if value == "" {
		return value
	}
	out := value
	for _, secret := range r.values {
		out = strings.ReplaceAll(out, secret, "[REDACTED:"+Fingerprint(secret)+"]")
	}
	out = jwtPattern.ReplaceAllStringFunc(out, func(token string) string {
		return "[REDACTED-JWT:" + Fingerprint(token) + "]"
	})
	return out
}

func (r Redactor) Headers(headers http.Header) map[string]string {
	out := map[string]string{}
	for key, values := range headers {
		joined := strings.Join(values, ",")
		if SensitiveName(key) {
			out[key] = "[REDACTED:" + Fingerprint(joined) + "]"
			continue
		}
		out[key] = r.String(joined)
	}
	return out
}

func (r Redactor) URL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return r.String(raw)
	}
	query := parsed.Query()
	for key := range query {
		if SensitiveName(key) {
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()
	return r.String(parsed.String())
}

func SensitiveName(name string) bool {
	normalized := strings.ToLower(name)
	return strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "api-key") ||
		strings.Contains(normalized, "apikey") ||
		strings.Contains(normalized, "api_key") ||
		normalized == "authorization" ||
		normalized == "cookie" ||
		normalized == "set-cookie"
}

func Fingerprint(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:12]
}
