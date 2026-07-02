// Package authn defines the authentication provider contract used by the HTTP
// layer, decoupling handlers from the concrete IAM client implementation.
package authn

import (
	"context"
	"errors"
)

// TokenSet is the result of a successful token exchange
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int // seconds until the access token expires
}

// User is the resolved identity behind an access token.
type User struct {
	Sub   string
	Name  string
	Email string
}

// Provider abstracts the IAM OAuth2/OIDC + user-lifecycle operations the blog
// needs (BFF pattern).
type Provider interface {
	// Login exchanges credentials for tokens via the password grant.
	Login(ctx context.Context, email, password string) (*TokenSet, error)
	// Register creates a self-service account (IAM sends a verification email).
	Register(ctx context.Context, email, password string) error
	// ForgotPassword triggers a reset email (always succeeds, anti-enumeration).
	ForgotPassword(ctx context.Context, email string) error
	// Logout revokes the refresh token.
	Logout(ctx context.Context, refreshToken string) error
	// Refresh exchanges a refresh token for a fresh token set.
	Refresh(ctx context.Context, refreshToken string) (*TokenSet, error)
	// UserInfo resolves identity claims for an access token.
	UserInfo(ctx context.Context, accessToken string) (*User, error)

	// FederatedLoginURL builds the URL that starts a federated (social) login for
	// the given provider (e.g. "google"). The browser is redirected here; the IdP
	// dance is handled by IAM, which then redirects back to redirectURI with an
	// authorization code carrying the passed-through state.
	FederatedLoginURL(provider, state, redirectURI string) string
	// ExchangeCode exchanges an authorization code (from the federated callback)
	// for a token set via the authorization_code grant.
	ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenSet, error)
}

var (
	// ErrInvalidCredentials is returned for bad username/password or grant errors.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUnavailable is returned when the IAM service cannot be reached.
	ErrUnavailable = errors.New("auth provider unavailable")
)
