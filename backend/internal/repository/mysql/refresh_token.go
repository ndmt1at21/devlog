package mysql

import (
	"context"
	"database/sql"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

type refreshTokenRepo struct{ db *sql.DB }

func (r *refreshTokenRepo) Create(ctx context.Context, t domain.RefreshToken) error {
	if t.ID == "" {
		t.ID = id.NewV7()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, token_hash, user_id, expires_at) VALUES (?,?,?,?)`,
		t.ID, t.TokenHash, t.UserID, t.ExpiresAt)
	return mapError(err)
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, token_hash, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = ?`, hash)
	var t domain.RefreshToken
	var revoked sql.NullTime
	if err := row.Scan(&t.ID, &t.TokenHash, &t.UserID, &t.ExpiresAt, &revoked); err != nil {
		return domain.RefreshToken{}, mapError(err)
	}
	if revoked.Valid {
		t.RevokedAt = revoked.Time
	}
	return t, nil
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, hash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`,
		timeNow(), hash)
	return mapError(err)
}
