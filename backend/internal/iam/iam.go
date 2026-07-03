// Package iam implements authn.Provider as a Backend-For-Frontend client of the
// IAM OAuth2/OIDC service. It uses the password grant for login, the
// user-lifecycle endpoints for register/forgot, and userinfo for identity.
// All calls target the tenant issuer base (e.g. http://localhost:8080/t/devnote).
package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
)

// Client is an IAM BFF client. It satisfies authn.Provider.
type Client struct {
	httpc        *http.Client
	issuer       string
	clientID     string
	clientSecret string
}

var _ authn.Provider = (*Client)(nil)

// New constructs a client against the tenant issuer base.
func New(issuer, clientID, clientSecret string) *Client {
	return &Client{
		httpc:        &http.Client{Timeout: 10 * time.Second},
		issuer:       strings.TrimRight(issuer, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

type tokenResp struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c *Client) token(ctx context.Context, form url.Values) (*authn.TokenSet, error) {
	form.Set("client_id", c.clientID)
	if c.clientSecret != "" {
		form.Set("client_secret", c.clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.issuer+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", authn.ErrUnavailable, err)
	}
	defer resp.Body.Close()

	var tr tokenResp
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &tr)

	if resp.StatusCode >= 400 || tr.AccessToken == "" {
		if tr.Error == "invalid_grant" || tr.Error == "invalid_client" || resp.StatusCode == http.StatusUnauthorized {
			return nil, authn.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("token endpoint status %d: %s %s", resp.StatusCode, tr.Error, tr.ErrorDescription)
	}
	return &authn.TokenSet{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		IDToken:      tr.IDToken,
		ExpiresIn:    tr.ExpiresIn,
	}, nil
}

// Login exchanges credentials via the password grant.
func (c *Client) Login(ctx context.Context, email, password string) (*authn.TokenSet, error) {
	return c.token(ctx, url.Values{
		"grant_type": {"password"},
		"username":   {email},
		"password":   {password},
		"scope":      {"openid profile email offline_access"},
	})
}

// Refresh exchanges a refresh token for a fresh token set.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*authn.TokenSet, error) {
	return c.token(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	})
}

// FederatedLoginURL builds the tenant's federated-login URL for a provider
// (google/facebook/oidc). IAM validates client_id + redirect_uri, runs the IdP
// flow, then redirects back to redirectURI with an IAM authorization code.
func (c *Client) FederatedLoginURL(provider, state, redirectURI string) string {
	q := url.Values{
		"client_id":    {c.clientID},
		"redirect_uri": {redirectURI},
		"state":        {state},
		"scope":        {"openid profile email offline_access"},
	}
	return c.issuer + "/auth/login/" + url.PathEscape(provider) + "?" + q.Encode()
}

// ExchangeCode swaps an authorization code for tokens (authorization_code grant).
// redirectURI must match the one used to start the flow.
func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*authn.TokenSet, error) {
	return c.token(ctx, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	})
}

// Register creates a self-service account; IAM emails a verification link.
func (c *Client) Register(ctx context.Context, email, password string) error {
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"clientId": c.clientID,
	})
	resp, err := c.postJSON(ctx, "/auth/register", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("%w", errConflict)
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

// ForgotPassword triggers a reset email. IAM always returns success.
func (c *Client) ForgotPassword(ctx context.Context, email string) error {
	payload, _ := json.Marshal(map[string]string{"email": email, "clientId": c.clientID})
	resp, err := c.postJSON(ctx, "/auth/forgot-password", payload)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Logout revokes the refresh token (best effort).
func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	form := url.Values{"token": {refreshToken}, "client_id": {c.clientID}}
	if c.clientSecret != "" {
		form.Set("client_secret", c.clientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.issuer+"/oauth2/revoke", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", authn.ErrUnavailable, err)
	}
	resp.Body.Close()
	return nil
}

// UserInfo resolves identity claims for an access token.
func (c *Client) UserInfo(ctx context.Context, accessToken string) (*authn.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.issuer+"/oauth2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", authn.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, authn.ErrInvalidCredentials
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("userinfo status %d", resp.StatusCode)
	}
	var claims map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, err
	}
	return &authn.User{
		Sub:   str(claims, "sub"),
		Email: str(claims, "email"),
		Name:  firstNonEmpty(str(claims, "name"), str(claims, "preferred_username"), str(claims, "username")),
	}, nil
}

// CheckPermissions asks IAM's policy decision endpoint whether the access token
// carries ALL of the required permissions. The PDP returns 200 with allow=false
// for invalid/expired tokens or missing permissions, so only transport/protocol
// failures surface as errors.
func (c *Client) CheckPermissions(ctx context.Context, accessToken string, required []string) (bool, error) {
	payload, _ := json.Marshal(map[string]any{
		"token":                accessToken,
		"required_permissions": required,
		"match_all":            true,
	})
	resp, err := c.postJSON(ctx, "/authz/decision", payload)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("authz decision status %d", resp.StatusCode)
	}
	var d struct {
		Allow  bool   `json:"allow"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return false, err
	}
	return d.Allow, nil
}

func (c *Client) postJSON(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.issuer+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", authn.ErrUnavailable, err)
	}
	return resp, nil
}

// errConflict signals a duplicate registration; handlers translate it.
var errConflict = fmt.Errorf("conflict")

// ErrConflict reports whether err is a duplicate-account error.
func ErrConflict(err error) bool { return err != nil && strings.Contains(err.Error(), "conflict") }

func str(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
