package domain

import (
	"context"
	"time"
)

// User roles. Permissions derive from the role (see the authlocal package);
// AuthorRole and AdminRole may create articles.
const (
	ReaderRole = "reader"
	AuthorRole = "author"
	AdminRole  = "admin"
)

// User is a blog account, stored locally (schema mirrors the IAM service's
// users table, single-tenant).
type User struct {
	ID            string
	Email         string // stored lowercased
	EmailVerified bool
	Name          string
	PasswordHash  string // argon2id PHC string; empty for federated-only accounts
	Role          string // ReaderRole | AuthorRole | AdminRole
	GoogleSub     string // stable Google subject when linked; empty otherwise
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RefreshToken is a stored (hashed) refresh token. The plaintext is never
// persisted; lookups use the SHA-256 hex hash.
type RefreshToken struct {
	ID        string
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	RevokedAt time.Time // zero when still valid
}

// Revoked reports whether the token has been revoked.
func (t RefreshToken) Revoked() bool { return !t.RevokedAt.IsZero() }

// UserRepository stores blog accounts.
type UserRepository interface {
	// Create persists a new user, assigning its ID. Returns ErrConflict when the
	// email (or linked Google subject) is already registered.
	Create(ctx context.Context, u User) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	// GetByEmail matches the lowercased email exactly.
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByGoogleSub(ctx context.Context, sub string) (User, error)
	// Update rewrites the user's mutable fields (name, hash, role, google_sub,
	// email_verified) and stamps UpdatedAt.
	Update(ctx context.Context, u User) error
	// Count returns the total number of accounts (used to bootstrap the first
	// registered user as author).
	Count(ctx context.Context) (int, error)
}

// RefreshTokenRepository stores hashed refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, t RefreshToken) error
	GetByHash(ctx context.Context, hash string) (RefreshToken, error)
	// Revoke marks the token revoked; revoking an unknown or already-revoked
	// token is a no-op.
	Revoke(ctx context.Context, hash string) error
}
