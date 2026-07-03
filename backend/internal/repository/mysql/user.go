package mysql

import (
	"context"
	"database/sql"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

type userRepo struct{ db *sql.DB }

const userCols = `id, email, email_verified, name, password_hash, role, google_sub, created_at, updated_at`

func scanUser(row *sql.Row) (domain.User, error) {
	var u domain.User
	var hash, googleSub sql.NullString
	if err := row.Scan(&u.ID, &u.Email, &u.EmailVerified, &u.Name, &hash, &u.Role, &googleSub, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return domain.User{}, mapError(err)
	}
	u.PasswordHash = hash.String
	u.GoogleSub = googleSub.String
	return u, nil
}

func (r *userRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	u.ID = id.NewV7()
	now := timeNow()
	u.CreatedAt, u.UpdatedAt = now, now
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, email_verified, name, password_hash, role, google_sub, created_at, updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		u.ID, u.Email, u.EmailVerified, u.Name, nullStr(u.PasswordHash), u.Role, nullStr(u.GoogleSub),
		u.CreatedAt, u.UpdatedAt); err != nil {
		return domain.User{}, mapError(err)
	}
	return u, nil
}

func (r *userRepo) GetByID(ctx context.Context, userID string) (domain.User, error) {
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE id = ?`, userID))
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE email = ?`, email))
}

func (r *userRepo) GetByGoogleSub(ctx context.Context, sub string) (domain.User, error) {
	if sub == "" {
		return domain.User{}, domain.ErrNotFound
	}
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE google_sub = ?`, sub))
}

func (r *userRepo) Update(ctx context.Context, u domain.User) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = ?, email_verified = ?, name = ?, password_hash = ?, role = ?, google_sub = ?, updated_at = ?
		 WHERE id = ?`,
		u.Email, u.EmailVerified, u.Name, nullStr(u.PasswordHash), u.Role, nullStr(u.GoogleSub), timeNow(), u.ID)
	if err != nil {
		return mapError(err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		// RowsAffected is 0 both for a missing row and a no-op update; confirm.
		if _, err := r.GetByID(ctx, u.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *userRepo) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return 0, mapError(err)
	}
	return n, nil
}
