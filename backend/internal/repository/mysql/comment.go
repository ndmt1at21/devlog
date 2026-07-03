package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

type commentRepo struct{ db *sql.DB }

func (r *commentRepo) ListByArticle(ctx context.Context, slug string) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, article_slug, name, body, created_at FROM comments
		 WHERE article_slug = ? ORDER BY created_at DESC`, slug)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.ArticleSlug, &c.Name, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *commentRepo) Create(ctx context.Context, c domain.Comment) (domain.Comment, error) {
	c.ID = id.NewV7()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO comments (id, article_slug, name, body, created_at) VALUES (?,?,?,?,?)`,
		c.ID, c.ArticleSlug, c.Name, c.Body, c.CreatedAt); err != nil {
		return domain.Comment{}, mapError(err)
	}
	return c, nil
}
