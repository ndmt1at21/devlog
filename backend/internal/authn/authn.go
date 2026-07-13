// Package authn defines the authentication provider contract used by the HTTP
// layer, decoupling handlers from the concrete provider implementation (the
// embedded authlocal provider).
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

// User is the resolved  identity behind an access token.
type User struct {
	Sub   string
	Name  string
	Email string
}

// Provider abstracts the auth + user-lifecycle operations the blog needs
// (BFF pattern).
type Provider interface {
	// Login exchanges credentials for tokens via the password grant.
	Login(ctx context.Context, email, password string) (*TokenSet, error)
	// Register creates a self-service account. Returns ErrConflict when the
	// email is already registered.
	Register(ctx context.Context, email, password, name string) error
	// ForgotPassword triggers a reset email (always succeeds, anti-enumeration).
	ForgotPassword(ctx context.Context, email string) error
	// Logout revokes the refresh token.
	Logout(ctx context.Context, refreshToken string) error
	// Refresh exchanges a refresh token for a fresh token set.
	Refresh(ctx context.Context, refreshToken string) (*TokenSet, error)
	// UserInfo resolves identity claims for an access token.
	UserInfo(ctx context.Context, accessToken string) (*User, error)

	// FederatedLoginURL builds the URL that starts a federated (social) login for
	// the given provider (e.g. "google"). The browser is redirected here to run
	// the IdP dance, which ends with a redirect back to redirectURI carrying an
	// authorization code and the passed-through state. Returns "" when the
	// provider is unknown or not configured.
	FederatedLoginURL(provider, state, redirectURI string) string
	// ExchangeCode exchanges an authorization code (from the federated callback)
	// for a token set via the authorization_code grant.
	ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenSet, error)

	// CheckPermissions reports whether the access token carries ALL of the
	// required permissions (names follow the "resource:action" convention),
	// mirroring IAM's policy decision semantics. An invalid/expired token
	// yields (false, nil); backend failures yield an error (callers fail closed).
	CheckPermissions(ctx context.Context, accessToken string, required []string) (bool, error)
}

var (
	// ErrInvalidCredentials is returned for bad username/password or grant errors.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUnavailable is returned when the auth backend cannot be reached.
	ErrUnavailable = errors.New("auth provider unavailable")
	// ErrConflict is returned by Register when the email is already registered.
	ErrConflict = errors.New("email already registered")
)
