// Package authlocal implements authn.Provider in-process, embedding the IAM
// service's core logic instead of proxying to it over HTTP: argon2id password
// hashing, signed access tokens, rotating hashed refresh tokens, role-based
// permission decisions, and Google federated login. Accounts live in the
// blog's own store (domain.Store), so no external auth service is required.
package authlocal

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

const (
	// accessTTL bounds how long a minted access token is honored.
	accessTTL = 15 * time.Minute
	// refreshTTL is the sliding window a session can stay alive without login;
	// each rotation issues a fresh window (matches the session cookie's 30 days).
	refreshTTL = 30 * 24 * time.Hour
	// refreshTokenBytes is the entropy of an opaque refresh token.
	refreshTokenBytes = 32
)

// rolePermissions maps a user role to its IAM-style "resource:action"
// permissions (the embedded equivalent of IAM's roles/permissions tables; the
// blog only distinguishes who may publish).
var rolePermissions = map[string][]string{
	domain.AuthorRole: {"articles:create"},
	domain.AdminRole:  {"articles:create"},
}

// Provider is the embedded auth provider. It satisfies authn.Provider.
type Provider struct {
	store              domain.Store
	jwtKey             []byte
	googleClientID     string
	googleClientSecret string
	httpc              *http.Client
}

var _ authn.Provider = (*Provider)(nil)

// GoogleConfig carries the Google OAuth client used for federated login.
// Empty ClientID disables the flow.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
}

// New builds the provider on top of the blog's store. secret keys the access
// tokens; it is domain-separated from other uses of the same secret.
func New(store domain.Store, secret string, google GoogleConfig) *Provider {
	key := sha256.Sum256([]byte("devlog-access-token:" + secret))
	return &Provider{
		store:              store,
		jwtKey:             key[:],
		googleClientID:     google.ClientID,
		googleClientSecret: google.ClientSecret,
		httpc:              &http.Client{Timeout: 10 * time.Second},
	}
}

// Login verifies the credentials against the stored argon2id hash and mints a
// token set (the embedded password grant).
func (p *Provider) Login(ctx context.Context, email, password string) (*authn.TokenSet, error) {
	u, err := p.store.Users().GetByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, domain.ErrNotFound) {
		// Burn the same argon2 work as a real verification so unknown emails
		// aren't distinguishable by response time.
		_, _ = hashPassword(password)
		return nil, authn.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	if u.PasswordHash == "" { // federated-only account
		return nil, authn.ErrInvalidCredentials
	}
	switch err := verifyPassword(password, u.PasswordHash); {
	case errors.Is(err, errMismatch):
		return nil, authn.ErrInvalidCredentials
	case err != nil:
		return nil, fmt.Errorf("verify password: %w", err)
	}
	return p.mintTokens(ctx, u)
}

// Register creates a self-service account. With no mail service embedded,
// accounts are active immediately (no verification email step).
func (p *Provider) Register(ctx context.Context, email, password, name string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u := domain.User{
		Email:         normalizeEmail(email),
		EmailVerified: true,
		Name:          strings.TrimSpace(name),
		PasswordHash:  hash,
		Role:          p.bootstrapRole(ctx),
	}
	if _, err := p.store.Users().Create(ctx, u); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return authn.ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// bootstrapRole makes the very first account an author so a fresh deployment
// can publish without manual role assignment; everyone after that is a reader.
func (p *Provider) bootstrapRole(ctx context.Context) string {
	if n, err := p.store.Users().Count(ctx); err == nil && n == 0 {
		return domain.AuthorRole
	}
	return domain.ReaderRole
}

// ForgotPassword always succeeds (anti-enumeration). There is no embedded mail
// service, so no reset email is sent; password resets are an operator task.
func (p *Provider) ForgotPassword(ctx context.Context, email string) error {
	return nil
}

// Logout revokes the refresh token (best effort).
func (p *Provider) Logout(ctx context.Context, refreshToken string) error {
	return p.store.RefreshTokens().Revoke(ctx, hashToken(refreshToken))
}

// Refresh rotates the refresh token and mints a fresh token set: the presented
// token is revoked and a new one with a full window is issued in its place.
func (p *Provider) Refresh(ctx context.Context, refreshToken string) (*authn.TokenSet, error) {
	hash := hashToken(refreshToken)
	t, err := p.store.RefreshTokens().GetByHash(ctx, hash)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, authn.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("load refresh token: %w", err)
	}
	if t.Revoked() || time.Now().After(t.ExpiresAt) {
		return nil, authn.ErrInvalidCredentials
	}
	u, err := p.store.Users().GetByID(ctx, t.UserID)
	if err != nil {
		return nil, authn.ErrInvalidCredentials
	}
	if err := p.store.RefreshTokens().Revoke(ctx, hash); err != nil {
		return nil, fmt.Errorf("rotate refresh token: %w", err)
	}
	return p.mintTokens(ctx, u)
}

// UserInfo resolves identity claims for an access token from the live user
// record (the embedded userinfo endpoint).
func (p *Provider) UserInfo(ctx context.Context, accessToken string) (*authn.User, error) {
	c, err := verifyJWT(p.jwtKey, accessToken)
	if err != nil {
		return nil, authn.ErrInvalidCredentials
	}
	u, err := p.store.Users().GetByID(ctx, c.Subject)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, authn.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	return &authn.User{Sub: u.ID, Email: u.Email, Name: u.Name}, nil
}

// CheckPermissions reports whether the access token's user holds ALL required
// permissions. Mirroring IAM's policy decision endpoint, an invalid or expired
// token is a plain deny (false, nil); only store failures surface as errors.
func (p *Provider) CheckPermissions(ctx context.Context, accessToken string, required []string) (bool, error) {
	c, err := verifyJWT(p.jwtKey, accessToken)
	if err != nil {
		return false, nil
	}
	// Decide from the live role, not the token snapshot, so grants and
	// revocations apply without waiting out the token TTL.
	u, err := p.store.Users().GetByID(ctx, c.Subject)
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("load user: %w", err)
	}
	granted := make(map[string]bool, len(rolePermissions[u.Role]))
	for _, perm := range rolePermissions[u.Role] {
		granted[perm] = true
	}
	for _, perm := range required {
		if !granted[perm] {
			return false, nil
		}
	}
	return true, nil
}

// FederatedLoginURL builds the IdP authorization URL the browser is redirected
// to. Only Google is wired; unknown providers or a missing Google client yield
// "" (callers treat that as "unavailable").
func (p *Provider) FederatedLoginURL(provider, state, redirectURI string) string {
	if provider != "google" || p.googleClientID == "" {
		return ""
	}
	return p.googleAuthCodeURL(state, redirectURI)
}

// ExchangeCode swaps the federated callback's authorization code for a local
// token set, creating or linking the account the external identity maps to.
func (p *Provider) ExchangeCode(ctx context.Context, code, redirectURI string) (*authn.TokenSet, error) {
	if p.googleClientID == "" {
		return nil, fmt.Errorf("%w: google login not configured", authn.ErrUnavailable)
	}
	ext, err := p.googleExchange(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}
	u, err := p.upsertFederated(ctx, ext)
	if err != nil {
		return nil, err
	}
	return p.mintTokens(ctx, u)
}

// upsertFederated resolves an external identity to a local user: match on the
// provider subject first, else link to an existing account with the same
// verified email, else create a fresh reader account (IAM's identity-linking
// rules, single-provider).
func (p *Provider) upsertFederated(ctx context.Context, ext *externalIdentity) (domain.User, error) {
	users := p.store.Users()

	u, err := users.GetByGoogleSub(ctx, ext.Subject)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, fmt.Errorf("load user by subject: %w", err)
	}

	if ext.Email != "" && ext.EmailVerified {
		u, err := users.GetByEmail(ctx, ext.Email)
		if err == nil {
			u.GoogleSub = ext.Subject
			u.EmailVerified = true
			if u.Name == "" {
				u.Name = ext.Name
			}
			if err := users.Update(ctx, u); err != nil {
				return domain.User{}, fmt.Errorf("link google identity: %w", err)
			}
			return u, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, fmt.Errorf("load user by email: %w", err)
		}
	}

	created, err := users.Create(ctx, domain.User{
		Email:         ext.Email,
		EmailVerified: ext.EmailVerified,
		Name:          ext.Name,
		Role:          p.bootstrapRole(ctx),
		GoogleSub:     ext.Subject,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create federated user: %w", err)
	}
	return created, nil
}

// mintTokens issues the access JWT and a stored, rotated refresh token.
func (p *Provider) mintTokens(ctx context.Context, u domain.User) (*authn.TokenSet, error) {
	now := time.Now()
	access, err := signJWT(p.jwtKey, claims{
		Issuer:      tokenIssuer,
		Subject:     u.ID,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(accessTTL).Unix(),
		Email:       u.Email,
		Name:        u.Name,
		Permissions: rolePermissions[u.Role],
	})
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refresh, refreshHash, err := generateOpaqueToken(refreshTokenBytes)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	if err := p.store.RefreshTokens().Create(ctx, domain.RefreshToken{
		TokenHash: refreshHash,
		UserID:    u.ID,
		ExpiresAt: now.Add(refreshTTL).UTC(),
	}); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &authn.TokenSet{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(accessTTL.Seconds()),
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
