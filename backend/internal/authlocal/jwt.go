package authlocal

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Access tokens are HS256 JWTs carrying the same claim shape as the IAM
// service's access tokens (iss/sub/exp/iat plus email, name and permissions).
// IAM signs with rotating RSA keys published on a JWKS endpoint; embedded in a
// single service there is no cross-service verification, so a symmetric key
// derived from the configured secret replaces the key-management machinery.

// tokenIssuer identifies tokens minted by this provider.
const tokenIssuer = "devlog"

// claims is the JWT payload for access tokens.
type claims struct {
	Issuer      string   `json:"iss"`
	Subject     string   `json:"sub"`
	IssuedAt    int64    `json:"iat"`
	ExpiresAt   int64    `json:"exp"`
	Email       string   `json:"email,omitempty"`
	Name        string   `json:"name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

var (
	errTokenInvalid = errors.New("token invalid")
	errTokenExpired = errors.New("token expired")
)

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// signJWT serializes and signs the claims with HMAC-SHA256.
func signJWT(key []byte, c claims) (string, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	signingInput := jwtHeader + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// verifyJWT checks the signature and expiry and returns the claims. It returns
// errTokenExpired for a well-signed but stale token and errTokenInvalid for
// everything else.
func verifyJWT(key []byte, raw string) (*claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 || parts[0] != jwtHeader {
		return nil, errTokenInvalid
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errTokenInvalid
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errTokenInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errTokenInvalid
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, errTokenInvalid
	}
	if c.Issuer != tokenIssuer {
		return nil, errTokenInvalid
	}
	if time.Now().Unix() >= c.ExpiresAt {
		return nil, errTokenExpired
	}
	return &c, nil
}

// generateOpaqueToken returns a random URL-safe token with n bytes of entropy
// and its SHA-256 hash (hex-encoded) for storage, mirroring IAM's refresh-token
// scheme: the plaintext is never persisted.
func generateOpaqueToken(n int) (plaintext, hash string, err error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", "", fmt.Errorf("read random: %w", err)
	}
	plaintext = base64.RawURLEncoding.EncodeToString(b)
	return plaintext, hashToken(plaintext), nil
}

// hashToken returns the hex-encoded SHA-256 hash of a token string.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
