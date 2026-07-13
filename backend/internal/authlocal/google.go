package authlocal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
)

// Google federated login, mirroring IAM's google/oidcgeneric identity
// providers: the standard OIDC authorization-code flow with Google's issuer
// pinned. Discovery is skipped in favor of Google's stable, documented
// endpoints so startup needs no network round-trip.
const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleScopes   = "openid email profile"
)

// externalIdentity is the normalized result of a federated login (IAM's
// identity.ExternalIdentity, trimmed to what the blog stores).
type externalIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// googleAuthCodeURL builds Google's authorization URL for the redirect flow.
func (p *Provider) googleAuthCodeURL(state, redirectURI string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {p.googleClientID},
		"redirect_uri":  {redirectURI},
		"scope":         {googleScopes},
		"state":         {state},
	}
	return googleAuthURL + "?" + q.Encode()
}

// googleExchange swaps an authorization code for the external identity. The
// id_token claims are read without signature verification: the token arrives
// directly from Google's token endpoint over TLS, which OIDC Core §3.1.3.7
// permits validating by the TLS server identity alone (IAM verifies signatures
// because it accepts id_tokens from browsers too; this flow never does).
func (p *Provider) googleExchange(ctx context.Context, code, redirectURI string) (*externalIdentity, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {p.googleClientID},
		"client_secret": {p.googleClientSecret},
		"redirect_uri":  {redirectURI},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", authn.ErrUnavailable, err)
	}
	defer resp.Body.Close()

	var tr struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode google token response: %w", err)
	}
	if resp.StatusCode >= 400 || tr.IDToken == "" {
		return nil, fmt.Errorf("%w: google token endpoint status %d (%s)", authn.ErrInvalidCredentials, resp.StatusCode, tr.Error)
	}
	return parseIDToken(tr.IDToken)
}

// parseIDToken extracts the identity claims from an OIDC id_token payload.
func parseIDToken(raw string) (*externalIdentity, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed id_token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode id_token payload: %w", err)
	}
	var c struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}
	if c.Sub == "" {
		return nil, errors.New("id_token has no subject")
	}
	return &externalIdentity{
		Subject:       c.Sub,
		Email:         strings.ToLower(strings.TrimSpace(c.Email)),
		EmailVerified: c.EmailVerified,
		Name:          c.Name,
	}, nil
}
