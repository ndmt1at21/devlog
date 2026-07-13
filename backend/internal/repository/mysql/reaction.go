package mysql

import (
	"context"
	"database/sql"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

type reactionRepo struct{ db *sql.DB }

func (r *reactionRepo) Set(ctx context.Context, kind domain.ReactionKind, slug, userID string, on bool) error {
	if !on {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM reactions WHERE article_slug = ? AND user_id = ? AND kind = ?`,
			slug, userID, string(kind))
		return mapError(err)
	}
	// Idempotent insert: re-liking/saving keeps the original created_at.
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO reactions (article_slug, user_id, kind, created_at) VALUES (?,?,?,?)
		 ON DUPLICATE KEY UPDATE created_at = created_at`,
		slug, userID, string(kind), timeNow())
	return mapError(err)
}

func (r *reactionRepo) Status(ctx context.Context, slug, userID string) (domain.ReactionStatus, error) {
	var st domain.ReactionStatus
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM reactions WHERE article_slug = ? AND kind = ?`,
		slug, string(domain.ReactionLike)).Scan(&st.Likes); err != nil {
		return domain.ReactionStatus{}, mapError(err)
	}
	if userID == "" {
		return st, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT kind FROM reactions WHERE article_slug = ? AND user_id = ?`, slug, userID)
	if err != nil {
		return domain.ReactionStatus{}, mapError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		if err := rows.Scan(&kind); err != nil {
			return domain.ReactionStatus{}, err
		}
		switch domain.ReactionKind(kind) {
		case domain.ReactionLike:
			st.Liked = true
		case domain.ReactionBookmark:
			st.Bookmarked = true
		}
	}
	return st, rows.Err()
}

func (r *reactionRepo) BookmarkedSlugs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT article_slug FROM reactions WHERE user_id = ? AND kind = ?
		 ORDER BY created_at DESC`, userID, string(domain.ReactionBookmark))
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		out = append(out, slug)
	}
	return out, rows.Err()
}
