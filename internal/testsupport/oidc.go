package testsupport

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type OIDCServer struct {
	Server   *httptest.Server
	Issuer   string
	ClientID string
	Key      *rsa.PrivateKey
	Kid      string
}

func NewOIDCServer(t *testing.T) *OIDCServer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	oidc := &OIDCServer{ClientID: "test-client", Key: key, Kid: "test-key"}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"issuer":                                oidc.Issuer,
			"authorization_endpoint":                oidc.Issuer + "/authorize",
			"token_endpoint":                        oidc.Issuer + "/token",
			"jwks_uri":                              oidc.Issuer + "/jwks",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{"keys": []any{oidc.jwk()}})
	})
	oidc.Server = httptest.NewServer(mux)
	oidc.Issuer = oidc.Server.URL
	t.Cleanup(oidc.Server.Close)
	return oidc
}

func (s *OIDCServer) IDToken(t *testing.T, nonce string, expiresAt time.Time, overrides map[string]any) string {
	t.Helper()
	claims := map[string]any{
		"iss":   s.Issuer,
		"sub":   "user-123",
		"aud":   s.ClientID,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
		"nonce": nonce,
	}
	for key, value := range overrides {
		claims[key] = value
	}
	return s.sign(t, claims)
}

func (s *OIDCServer) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "RS256", "typ": "JWT", "kid": s.Kid}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	sum := sha256.Sum256([]byte(unsigned))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.Key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (s *OIDCServer) jwk() map[string]string {
	pub := s.Key.Public().(*rsa.PublicKey)
	return map[string]string{
		"kty": "RSA",
		"use": "sig",
		"kid": s.Kid,
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
